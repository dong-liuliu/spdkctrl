/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

// Package spdk adds support for the TEST_SPDK_VHOST_SOCKET env variable to test binaries
// and manages the SPDK instance for tests.
package spdk_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nightlyone/lockfile"
	"github.com/pkg/errors"

	"github.com/intel/oim/pkg/log"
	"github.com/intel/oim/pkg/oim-common"
	"github.com/intel/oim/pkg/spdk"
	// . "github.com/onsi/ginkgo"
)

var (
	// SPDK handle connected to the running SPDK.
	SPDK *spdk.Client
	// SPDKPath is the path to the socket of the running SPDK.
	SPDKPath string
	// VHostPath is the vhost socket for the SCSI VHost controller of the running SPDK.
	VHostPath string

	// VHost controller name.
	VHost = "e2e-test-vhost"

	// VHostDev is the BDF string for the SCSI VHost controller
	VHostDev = "0000:00:15.0"

	spdkSock = os.Getenv("TEST_SPDK_VHOST_SOCKET")
	spdkApp  = os.Getenv("TEST_SPDK_VHOST_BINARY")
	lock     *lockfile.Lockfile
	spdkCmd  *exec.Cmd
	tmpDir   string
	spdkOut  io.WriteCloser

	o opts
)

type opts struct {
	controller bool
	socket     string
}

// Option is the argument type for Init.
type Option func(*opts)

// WithVHostSCSI enables the creation of a SCSI controller.
func WithVHostSCSI() Option {
	return func(o *opts) {
		o.controller = true
	}
}

// WithSPDKSocket overrides the default env variables and
// causes Init to connect to an existing daemon, without
// locking it for exclusive use. This is meant to be used
// in a parallel Gingko test run where the master node
// does the normal Init and the rest of the nodes
// use the socket.
func WithSPDKSocket(path string) Option {
	return func(o *opts) {
		o.socket = path
	}
}

// Init connects to SPDK and creates a VHost SCSI controller.
// Must be matched by a Finalize call, even after a failure.
func Init(options ...Option) error {
	for _, op := range options {
		op(&o)
	}

	// Connect to existing SPDK?
	if o.socket != "" {
		s, err := spdk.New(o.socket)
		if err != nil {
			return err
		}
		SPDK = s
		SPDKPath = o.socket
		return nil
	}

	// Set up VHost SCSI, if we have SPDK.
	if spdkSock == "" && spdkApp == "" {
		return nil
	}

	if SPDK != nil || VHostPath != "" || spdkCmd != nil {
		return errors.New("Finalize not called or failed")
	}

	if spdkApp != "" {
		// TODO: suppress logging to syslog
		t, err := ioutil.TempDir("", "spdk")
		if err != nil {
			return errors.Wrap(err, "SPDK temp directory")
		}
		tmpDir = t
		spdkSock = filepath.Join(tmpDir, "spdk.sock")
		spdkOut = oimcommon.LogWriter(log.L().With("at", "spdk"))
		var done <-chan interface{}
		{
			log.L().Infof("Starting %s", spdkApp)
			cmd := exec.Command("sudo", spdkApp, "-R", "-S", tmpDir, "-r", spdkSock,
				// Use less precious huge pages. 64MB
				// and 128MB are not enough and cause
				// out-of-memory errors for various
				// allocations during startup. With
				// the default of HUGEMEM=2048 that
				// means that we can start 8 instances
				// in parallel, and four in parallel
				// with a VM of 1GB. If testing fails
				// when run in parallel, then more
				// huge pages need to be reserved.
				"-s", "256",
			)
			// Start with its own process group so that we can kill sudo
			// and its child spdkApp via the process group.
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			cmd.Stdout = spdkOut
			cmd.Stderr = spdkOut
			cm, err := oimcommon.AddCmdMonitor(cmd)
			if err != nil {
				return errors.Wrap(err, "monitor command")
			}
			if err := cmd.Start(); err != nil {
				return err
			}
			done = cm.Watch()
			spdkCmd = cmd
		}
		// Starting up can be slow when the number of reserved huge pages is high or
		// many processes are running.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	loop:
		for {
			select {
			case <-done:
				return errors.New("SPDK quit unexpectedly")

			case <-ctx.Done():
				return fmt.Errorf("Timed out waiting for %s", spdkSock)

			case <-time.After(time.Millisecond):
				_, err := os.Stat(spdkSock)
				if err == nil {
					break loop
				}
			}
		}
		{
			cmd := exec.CommandContext(ctx, "sudo", "chmod", "a+rw", spdkSock) // nolint: gosec
			out, err := cmd.CombinedOutput()
			if err != nil {
				return errors.Wrapf(err, "chmod %s: %s", spdkSock, out)
			}
		}

	}

	// Protect against other processes using the same daemon.
	l, err := lockfile.New(spdkSock + ".testlock")
	if err == nil {
		for {
			err = l.TryLock()
			if te, ok := err.(interface{ Temporary() bool }); !ok || !te.Temporary() {
				break
			}
			time.Sleep(time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("Locking %s.testlock: %s", spdkSock, err)
	}
	lock = &l

	s, err := spdk.New(spdkSock)
	if err != nil {
		return err
	}
	SPDK = s
	SPDKPath = spdkSock

	if o.controller {
		args := spdk.ConstructVHostSCSIControllerArgs{
			Controller: VHost,
		}
		err = spdk.ConstructVHostSCSIController(context.Background(), SPDK, args)
		if err != nil {
			return err
		}
		VHostPath = filepath.Join(filepath.Dir(spdkSock), VHost)

		// If we are not running as root, we need to
		// change permissions on the new socket.
		if os.Getuid() != 0 {
			cmd := exec.Command("sudo", "chmod", "a+rw", VHostPath) // nolint: gosec
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("'sudo chmod' on vhost socket %s failed: %s\n%s", VHostPath, err, string(out))
			}
		}
	} else {
		VHostPath = ""
	}

	return nil
}

// Finalize frees any resources allocated by Init. Safe to call without
// Init or after Init failure.
func Finalize() error {
	if SPDK != nil {
		if VHostPath != "" {
			args := spdk.RemoveVHostControllerArgs{
				Controller: VHost,
			}
			// We try to clean up, but that can fail when someone left a disk attached
			// to the controller ("Trying to remove non-empty controller").
			// Just log such errors and proceed, as we'll kill the process anyway.
			log.L().Infof("Removing VHost SCSI controller %s", VHost)
			if err := spdk.RemoveVHostController(context.Background(), SPDK, args); err != nil {
				log.L().Errorw("RemoveVHostController failed", "error", err)
			}
			VHostPath = ""
		}
		if err := SPDK.Close(); err != nil {
			log.L().Errorw("close SPDK socket", "error", err)
		}
		SPDK = nil
	}
	if spdkCmd != nil {
		// Kill the process group to catch both child (sudo) and grandchild (SPDK).
		timer := time.AfterFunc(30*time.Second, func() {
			log.L().Infof("Killing SPDK vhost %d", spdkCmd.Process.Pid)
			exec.Command("sudo", "--non-interactive", "kill", "-9", fmt.Sprintf("-%d", spdkCmd.Process.Pid)).CombinedOutput() // nolint: gosec
		})
		defer timer.Stop()
		log.L().Infof("Stopping SPDK vhost %d", spdkCmd.Process.Pid)
		exec.Command("sudo", "--non-interactive", "kill", fmt.Sprintf("-%d", spdkCmd.Process.Pid)).CombinedOutput() // nolint: gosec
		spdkCmd.Wait()                                                                                              // nolint: gosec
		spdkCmd = nil
	}
	if lock != nil {
		if err := lock.Unlock(); err != nil {
			return err
		}
	}
	if spdkOut != nil {
		if err := spdkOut.Close(); err != nil {
			return err
		}
	}
	if tmpDir != "" {
		if err := os.RemoveAll(tmpDir); err != nil {
			return err
		}
	}
	return nil
}
