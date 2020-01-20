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

type NbdStartDiskArgs struct {
	BdevName  string `json:"bdev_name"`
	NbdDevice string `json:"nbd_device"`
}

//NbdStartDiskResponse is string: path of exported Nbd disk.
func NbdStartDisk(ctx context.Context, client *Client, args NbdStartDiskArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "nbd_start_disk", args, &response)
	if err != nil {
		return "", err
	}
	return response, nil
}

type NbdGetDisksArgs struct {
	NbdDevice string `json:"nbd_device,omitempty"`
}

type NbdGetDisksResponse []NbdStartDiskArgs

func NbdGetDisks(ctx context.Context, client *Client, args NbdGetDisksArgs) (NbdGetDisksResponse, error) {
	var response NbdGetDisksResponse
	err := client.Invoke(ctx, "nbd_get_disks", args, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type NbdStopDiskArgs struct {
	NbdDevice string `json:"nbd_device"`
}

//NbdStopDiskResponse is "bool": indication of result
func NbdStopDisk(ctx context.Context, client *Client, args NbdStopDiskArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "nbd_stop_disk", args, &response)
	if err != nil {
		return false, err
	}
	return response, nil
}
