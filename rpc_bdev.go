/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

// Package spdkctrl provides Go bindings for the SPDK JSON 2.0 RPC interface
// (http://www.spdk.io/doc/jsonrpc.html).
package spdkctrl

import (
	"context"
)

type SupportedIOTypes struct {
	Read       bool `json:"read"`
	Write      bool `json:"write"`
	Unmap      bool `json:"unmap"`
	WriteZeros bool `json:"write_zeroes"`
	Flush      bool `json:"flush"`
	Reset      bool `json:"reset"`
	NVMEAdmin  bool `json:"nvme_admin"`
	NVMEIO     bool `json:"nvme_io"`
}

type Bdev struct {
	Name             string           `json:"name"`
	ProductName      string           `json:"product_name"`
	UUID             string           `json:"uuid"`
	BlockSize        int64            `json:"block_size"`
	NumBlocks        int64            `json:"num_blocks"`
	Claimed          bool             `json:"claimed"`
	Zoned            bool             `json:"zoned"`
	SupportedIOTypes SupportedIOTypes `json:"supported_io_types"`
	DriverSpecific   *interface{}     `json:"driver_specific"`
}

type BdevGetBdevsArgs struct {
	Name string `json:"name,omitempty"`
}

type BdevGetBdevsResponse []Bdev

func BdevGetBdevs(ctx context.Context, client *Client, args BdevGetBdevsArgs) (BdevGetBdevsResponse, error) {
	var response BdevGetBdevsResponse
	err := client.Invoke(ctx, "bdev_get_bdevs", args, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type BdevMallocCreateArgs struct {
	Name      string `json:"name,omitempty"`
	BlockSize int64  `json:"block_size"`
	NumBlocks int64  `json:"num_blocks"`
	UUID      string `json:"uuid,omitempty"`
}

//BdevMallocCreateResponse is "string": name of newly created bdev
func BdevMallocCreate(ctx context.Context, client *Client, args BdevMallocCreateArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "bdev_malloc_create", args, &response)
	if err != nil {
		return "", err
	}
	return response, err
}

type BdevMallocDeleteArgs struct {
	Name string `json:"name"`
}

//BdevMallocDeleteResponse is "bool": indication of delete result
func BdevMallocDelete(ctx context.Context, client *Client, args BdevMallocDeleteArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "bdev_malloc_delete", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}

type BdevAioCreateArgs struct {
	Name      string `json:"name"`
	Filename  string `json:"filename"`
	BlockSize int64  `json:"block_size,omitempty"`
}

//BdevAioCreateResponse is "string": name of newly created bdev
func BdevAioCreate(ctx context.Context, client *Client, args BdevAioCreateArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "bdev_aio_create", args, &response)
	if err != nil {
		return "", err
	}
	return response, err
}

type BdevAioDeleteArgs struct {
	Name string `json:"name"`
}

//BdevAioDeleteResponse is "bool": indication of delete result
func BdevAioDelete(ctx context.Context, client *Client, args BdevAioDeleteArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "bdev_aio_delete", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}
