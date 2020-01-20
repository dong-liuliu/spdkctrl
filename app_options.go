/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

package spdkctrl

import (
	"os"
)

type appOpts struct {
	spdkApp       string
	appSocket     string
	vhostSockPath string
	// SPDK app output
	logOutput *os.File
}

// AppOption is the argument type for AppStart.
type AppOption func(*appOpts)

func WithSpdkApp(path string) AppOption {
	return func(o *appOpts) {
		o.spdkApp = path
	}
}

func WithVhostSockPath(path string) AppOption {
	return func(o *appOpts) {
		o.vhostSockPath = path
	}
}

// WithSPDKSocket overrides the default env variables and
// causes Init to connect to an existing daemon, without
// locking it for exclusive use. This is meant to be used
// in a parallel Gingko test run where the master node
// does the normal Init and the rest of the nodes
// use the socket.
func WithAppSocket(path string) AppOption {
	return func(o *appOpts) {
		o.appSocket = path
	}
}

func WithLogOutput(out *os.File) AppOption {
	return func(o *appOpts) {
		o.logOutput = out
	}
}

func appOptions2Args(opts *appOpts) []string {
	appArgs := []string{}

	appArgs = append(appArgs, opts.spdkApp)

	if opts.appSocket != "" {
		appArgs = append(appArgs, "-r", opts.appSocket)
	}

	if opts.vhostSockPath != "" {
		appArgs = append(appArgs, "-S", opts.vhostSockPath)
	}

	return appArgs
}
