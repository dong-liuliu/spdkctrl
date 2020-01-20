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

func appInit(t *testing.T) *spdk.App {
	optsFunc := []spdk.AppOption{}

	optsFunc = append(optsFunc, spdk.WithSpdkApp(testSpdkApp))
	optsFunc = append(optsFunc, spdk.WithAppSocket(testSpdkAppSocket))
	optsFunc = append(optsFunc, spdk.WithVhostSockPath(testSpdkVhostsockPath))
	optsFunc = append(optsFunc, spdk.WithLogOutput(os.Stdout))

	spdkApp, err := spdk.AppRun(optsFunc...)
	assert.NoError(t, err, "Failed to Run SPDK app: %s", err)
	assert.NotEmpty(t, spdkApp, "SPDK app is nil")

	return spdkApp
}

func appFini(t *testing.T, spdkApp *spdk.App) {
	forced := spdk.AppTerm(spdkApp, true)
	assert.False(t, forced, "SPDK app is terminated by force")
}

func TestClient(t *testing.T) {
	spdkApp := appInit(t)
	defer appFini(t, spdkApp)

	client, err := spdk.NewClient(testSpdkAppSocket, os.Stdout)
	assert.NoError(t, err, "Failed to connect SPDK app socket %s: %s", testSpdkAppSocket, err)
	assert.NotEmpty(t, client, "SPDK client is nil")

	err = client.Close()
	assert.NoError(t, err, "Failed to close SPDK client: %s", err)
}
