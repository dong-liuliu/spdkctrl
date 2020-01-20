/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

package spdkctrl_test

import (
	"context"
	"fmt"

	"path"
	"testing"

	spdk "github.com/dong-liuliu/spdkctrl"
	"github.com/stretchr/testify/assert"
)

func connect(t *testing.T) *spdk.Client {
	client, err := spdk.NewClient(testSpdkAppSocket, nil)
	assert.NoError(t, err, "Failed to connect SPDK app socket %s: %s", testSpdkAppSocket, err)
	assert.NotEmpty(t, client, "SPDK client is nil")

	return client
}

func disconnect(t *testing.T, spdkClient *spdk.Client) {
	err := spdkClient.Close()
	assert.NoError(t, err, "Failed to close SPDK client: %s", err)
}

func testRpcBdevMallocCreate(t *testing.T, spdkClient *spdk.Client) {
	argBdevMallocCreate := spdk.BdevMallocCreateArgs{
		NumBlocks: 10240,
		BlockSize: 4096,
	}

	response, err := spdk.BdevMallocCreate(context.Background(), spdkClient, argBdevMallocCreate)
	assert.NoError(t, err, "Failed to create malloc bdev: %s", err)
	assert.NotEmpty(t, response, "Unexpected empty bdev name")
	fmt.Println("Created Malloc bdev:", response)
}

func testRpcBdevMallocDelete(t *testing.T, spdkClient *spdk.Client) {
	argBdevMallocDelete := spdk.BdevMallocDeleteArgs{
		Name: "Malloc0",
	}

	response, err := spdk.BdevMallocDelete(context.Background(), spdkClient, argBdevMallocDelete)
	assert.NoError(t, err, "Failed to delete malloc bdev: %s", err)
	assert.True(t, response, "Failed to delete malloc bdev: %b", response)
	fmt.Println("Deleted:", response)
}

func testRpcBdevGetBdevs(t *testing.T, spdkClient *spdk.Client, is_nil bool) {
	response, err := spdk.BdevGetBdevs(context.Background(), spdkClient, spdk.BdevGetBdevsArgs{})
	assert.NoError(t, err, "Failed to list bdevs: %s", err)
	if is_nil {
		assert.Empty(t, response, "Unexpected non-empty bdev list")
	} else {
		assert.NotEmpty(t, response, "Unexpected empty bdev list")
	}

	fmt.Println("Gotten bdevs are:", response)
}

func TestRpcMallocBdev(t *testing.T) {
	spdkApp := appInit(t)
	defer appFini(t, spdkApp)

	spdkClient := connect(t)
	defer disconnect(t, spdkClient)

	// At start, there is no bdev, so response is an empty bdev list
	testRpcBdevGetBdevs(t, spdkClient, true)

	testRpcBdevMallocCreate(t, spdkClient)

	testRpcBdevGetBdevs(t, spdkClient, false)

	testRpcBdevMallocDelete(t, spdkClient)

	testRpcBdevGetBdevs(t, spdkClient, true)
}

func testRpcBdevAioCreate(t *testing.T, spdkClient *spdk.Client) {
	argBdevAioCreate := spdk.BdevAioCreateArgs{
		Name:      "Aio0",
		Filename:  path.Join(testSpdkVhostsockPath, "aiodisk"),
		BlockSize: 4096,
	}

	response, err := spdk.BdevAioCreate(context.Background(), spdkClient, argBdevAioCreate)
	assert.NoError(t, err, "Failed to create Aio bdev: %s", err)
	assert.NotEmpty(t, response, "Unexpected empty bdev name")
	fmt.Println("Created Aio bdev:", response)
}

func testRpcBdevAioDelete(t *testing.T, spdkClient *spdk.Client) {
	argBdevAioDelete := spdk.BdevAioDeleteArgs{
		Name: "Aio0",
	}

	response, err := spdk.BdevAioDelete(context.Background(), spdkClient, argBdevAioDelete)
	assert.NoError(t, err, "Failed to delete Aio bdev: %s", err)
	assert.True(t, response, "Failed to delete Aio bdev: %b", response)
	fmt.Println("Deleted:", response)
}

func TestRpcAioBdev(t *testing.T) {
	spdkApp := appInit(t)
	defer appFini(t, spdkApp)

	spdkClient := connect(t)
	defer disconnect(t, spdkClient)

	// At start, there is no bdev, so response is an empty bdev list
	testRpcBdevGetBdevs(t, spdkClient, true)

	testRpcBdevAioCreate(t, spdkClient)

	testRpcBdevGetBdevs(t, spdkClient, false)

	testRpcBdevAioDelete(t, spdkClient)

	testRpcBdevGetBdevs(t, spdkClient, true)
}
