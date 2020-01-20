/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

package spdkctrl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/dong-liuliu/spdkctrl/utils"
	log "github.com/sirupsen/logrus"
)

const (
	SpdkAppTimeout = 10
)

type App struct {
	spdkCmd *exec.Cmd
	logger  *log.Logger
}

func appReady(spdkApp *App, spdkAppSocket string) error {
	cmd := spdkApp.spdkCmd
	cm, err := utils.AddCmdMonitor(cmd)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	// Starting up can be slow when the number of reserved huge pages is high or
	// many processes are running.
	ctx, cancel := context.WithTimeout(context.Background(), SpdkAppTimeout*time.Second)
	defer cancel()
	done := cm.Watch()

	for {
		select {

		case <-done:
			return fmt.Errorf("SPDK quit unexpectedly")

		case <-ctx.Done():
			logger := spdkApp.logger
			logger.Infof("Killing SPDK vhost %d", cmd.Process.Pid)
			exec.Command("sudo", "--non-interactive", "kill", "-9", fmt.Sprintf("-%d", cmd.Process.Pid)).CombinedOutput()
			return fmt.Errorf("Timed out waiting for %s", spdkAppSocket)

		case <-time.After(time.Millisecond):
			_, err := os.Stat(spdkAppSocket)
			if err == nil {
				return nil
			}
		}
	}

	return nil
}

func AppRun(options ...AppOption) (*App, error) {
	var spdkApp App
	var opts appOpts
	var err error

	spdkAppBinary := os.Getenv("SPDK_APP_BINARY")
	spdkAppSocket := os.Getenv("SPDK_APP_SOCKET")
	spdkVhostSocketPath := os.Getenv("SPDK_VHOST_SOCKET_PATH")

	opts.spdkApp = spdkAppBinary
	opts.appSocket = spdkAppSocket
	opts.vhostSockPath = spdkVhostSocketPath

	// Get user specific options for application
	for _, op := range options {
		op(&opts)
	}

	logger := log.New()
	logger.Out = opts.logOutput
	spdkApp.logger = logger

	if opts.spdkApp == "" {
		err = errors.New("SPDK application is not assigned")
		logger.Errorln(err)
		return nil, err
	}

	// Adopt SPDK app options
	appArgs := appOptions2Args(&opts)

	logger.Infoln("Starting app", appArgs)
	cmd := exec.Command("sudo", appArgs[:]...)
	spdkApp.spdkCmd = cmd

	// Start with its own process group so that we can kill sudo
	// and its child spdkApp via the process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = opts.logOutput
	cmd.Stderr = opts.logOutput

	err = appReady(&spdkApp, opts.appSocket)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	logger.Infoln("App is ready")

	spdkApp.spdkCmd = cmd

	{
		ctx, cancel := context.WithTimeout(context.Background(), SpdkAppTimeout*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sudo", "chmod", "a+rw", opts.appSocket)
		out, err := cmd.CombinedOutput()
		if err != nil {
			err = errors.New(fmt.Sprintln(err.Error(), "chmod %s: %s", opts.appSocket, out))
			logger.Errorln(err)
			return nil, err
		}
	}

	return &spdkApp, nil
}

func AppTerm(spdkApp *App, force bool) bool {
	forced := false

	if spdkApp == nil || spdkApp.spdkCmd == nil {
		return false
	}

	spdkCmd := spdkApp.spdkCmd
	logger := spdkApp.logger

	// Kill the process group to catch both child (sudo) and grandchild (SPDK).
	if force {
		timer := time.AfterFunc(SpdkAppTimeout*time.Second, func() {
			logger.Infof("Killing SPDK vhost %d", spdkCmd.Process.Pid)
			exec.Command("sudo", "--non-interactive", "kill", "-9", fmt.Sprintf("-%d", spdkCmd.Process.Pid)).CombinedOutput()
			forced = true
		})
		defer timer.Stop()
	}

	logger.Infof("Stopping SPDK vhost %d", spdkCmd.Process.Pid)
	cmd := exec.Command("sudo", "--non-interactive", "kill", fmt.Sprintf("%d", spdkCmd.Process.Pid))
	_, err := cmd.CombinedOutput()
	if err == nil {
		cmd.Wait()
	} else {
		spdkCmd.Wait()
	}

	spdkApp.spdkCmd = nil
	logger.Infof("Stopped SPDK vhost %d", spdkCmd.Process.Pid)

	return forced
}
