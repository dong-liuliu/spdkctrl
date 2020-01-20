/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

package spdkctrl_test

import (
	"os"
	"testing"

	spdk "github.com/dong-liuliu/spdkctrl"
	"github.com/stretchr/testify/assert"
)

const (
	testSpdkApp           = "/root/go/src/github.com/spdk/spdk/app/vhost/vhost"
	testSpdkAppSocket     = "/tmp/spdk/spdk.sock"
	testSpdkVhostsockPath = "/tmp/spdk"
)

func TestAppByOptions(t *testing.T) {

	optsFunc := []spdk.AppOption{}

	optsFunc = append(optsFunc, spdk.WithSpdkApp(testSpdkApp))
	optsFunc = append(optsFunc, spdk.WithAppSocket(testSpdkAppSocket))
	optsFunc = append(optsFunc, spdk.WithVhostSockPath(testSpdkVhostsockPath))
	optsFunc = append(optsFunc, spdk.WithLogOutput(os.Stdout))

	spdkApp, err := spdk.AppRun(optsFunc...)
	assert.NoError(t, err, "Failed to Run SPDK app: %s", err)
	assert.NotEmpty(t, spdkApp, "SPDK app is nil")

	forced := spdk.AppTerm(spdkApp, true)
	assert.False(t, forced, "SPDK app is terminated by force")
}

func TestAppByEnv(t *testing.T) {
	optsFunc := []spdk.AppOption{}

	optsFunc = append(optsFunc, spdk.WithLogOutput(os.Stdout))

	os.Setenv("SPDK_APP_BINARY", testSpdkApp)
	os.Setenv("SPDK_APP_SOCKET", testSpdkAppSocket)
	os.Setenv("SPDK_VHOST_SOCKET_PATH", testSpdkVhostsockPath)

	spdkApp, err := spdk.AppRun(optsFunc...)
	assert.NoError(t, err, "Failed to Run SPDK app: %s", err)
	assert.NotEmpty(t, spdkApp, "SPDK app is nil")

	forced := spdk.AppTerm(spdkApp, true)
	assert.False(t, forced, "SPDK app is terminated by force")
}
