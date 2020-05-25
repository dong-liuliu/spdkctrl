/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

// Package spdkctrl provides Go bindings for the SPDK JSON 2.0 RPC interface
// (https://spdk.io/doc/jsonrpc.html#jsonrpc_components_lvol).
package spdkctrl

import (
	"context"
	"fmt"
)

type BdevLvolCreateLvstoreArgs struct {
	BdevName  string `json:"bdev_name"`
	LvsName   string `json:"lvs_name"`
	ClusterSz int64  `json:"cluster_sz,omitempty"`
	// clear method for data region. Available: none, unmap (default), write_zeroes
	ClearMethod string `json:"clear_method,omitempty"`
}

//BdevLvolCreateLvstoreResponse is "string": UUID of the created logical volume store
func BdevLvolCreateLvstore(ctx context.Context, client *Client, args BdevLvolCreateLvstoreArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "bdev_lvol_create_lvstore", args, &response)
	if err != nil {
		return "", err
	}
	return response, err
}

type BdevLvolDeleteLvstoreArgs struct {
	//Either uuid or lvs_name must be specified, but not both.
	Uuid    string `json:"uuid,omitempty"`
	LvsName string `json:"lvs_name,omitempty"`
}

//BdevLvolDeleteLvstoreResponse is "bool": indication of delete result
func BdevLvolDeleteLvstore(ctx context.Context, client *Client, args BdevLvolDeleteLvstoreArgs) (bool, error) {
	var response bool

	if args.LvsName == "" && args.Uuid == "" {
		return false, fmt.Errorf("invalid parameters")
	}
	if args.LvsName != "" && args.Uuid != "" {
		return false, fmt.Errorf("invalid parameters")
	}

	err := client.Invoke(ctx, "bdev_lvol_delete_lvstore", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}

type BdevLvolGetLvstoresArgs struct {
	//Either uuid or lvs_name may be specified, but not both.
	//If both uuid and lvs_name are omitted, information about
	//all logical volume stores is returned
	Uuid    string `json:"uuid,omitempty"`
	LvsName string `json:"lvs_name,omitempty"`
}

type Lvstore struct {
	Uuid              string `json:"uuid"`
	BaseBdev          string `json:"base_bdev"`
	FreeClusters      int    `json:"free_clusters"`
	ClusterSize       int    `json:"cluster_size"`
	TotalDataClusters int    `json:"total_data_clusters"`
	BlockSize         int    `json:"block_size"`
	Name              string `json:"name"`
}

type BdevLvolGetLvstoresResponse []Lvstore

func BdevLvolGetLvstores(ctx context.Context, client *Client, args BdevLvolGetLvstoresArgs) (BdevLvolGetLvstoresResponse, error) {
	var response BdevLvolGetLvstoresResponse

	if args.LvsName != "" && args.Uuid != "" {
		return nil, fmt.Errorf("invalid parameters")
	}

	err := client.Invoke(ctx, "bdev_lvol_get_lvstores", args, &response)
	if err != nil {
		return nil, err
	}
	return response, err
}

type BdevLvolCreateArgs struct {
	//lvol_name will be used in the alias of the created logical volume.
	LvolName string `json:"lvol_name"`
	//Size will be rounded up to a multiple of cluster size.
	Size          int64 `json:"size"`
	ThinProvision bool  `json:"thin_provision,omitempty"`
	//Either uuid or lvs_name must be specified, but not both.
	Uuid        string `json:"uuid,omitempty"`
	LvsName     string `json:"lvs_name,omitempty"`
	ClearMethod string `json:"clear_method,omitempty"`
}

//BdevLvolCreateResponse is "string": UUID of the created logical volume is returned.
func BdevLvolCreate(ctx context.Context, client *Client, args BdevLvolCreateArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "bdev_lvol_create", args, &response)
	if err != nil {
		return "", err
	}
	return response, err
}

type BdevLvolDeleteArgs struct {
	//UUID or alias of the logical volume to destroy
	Name string `json:"name"`
}

//BdevLvolDeleteResponse is "bool": indication of delete result
func BdevLvolDelete(ctx context.Context, client *Client, args BdevLvolDeleteArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "bdev_lvol_delete", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}

type BdevLvolSnapshotArgs struct {
	//UUID or alias of the logical volume to snapshot
	LvolName     string `json:"lvol_name"`
	SnapshotName string `json:"snapshot_name"`
}

//BdevLvolSnapshotResponse is "string": UUID of the created logical volume snapshot is returned.
func BdevLvolSnapshot(ctx context.Context, client *Client, args BdevLvolSnapshotArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "bdev_lvol_snapshot", args, &response)
	if err != nil {
		return "", err
	}
	return response, err
}

type BdevLvolCloneArgs struct {
	//UUID or alias of the logical volume to clone
	SnapshotName string `json:"snapshot_name"`
	CloneName    string `json:"clone_name"`
}

//BdevLvolCloneResponse is "string": UUID of the created logical volume clone is returned.
func BdevLvolClone(ctx context.Context, client *Client, args BdevLvolCloneArgs) (string, error) {
	var response string
	err := client.Invoke(ctx, "bdev_lvol_clone", args, &response)
	if err != nil {
		return "", err
	}
	return response, err
}

type BdevLvolSetReadOnlyArgs struct {
	//UUID or alias of the logical volume to set
	Name string `json:"name"`
}

//BdevLvolSetReadOnlyResponse is "bool": result
func BdevLvolSetReadOnly(ctx context.Context, client *Client, args BdevLvolSetReadOnlyArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "bdev_lvol_set_read_only", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}

type BdevLvolDecoupleParentArgs struct {
	//UUID or alias of the logical volume to set
	Name string `json:"name"`
}

//BdevLvolDecoupleParentResponse is "bool": result
func BdevLvolDecoupleParent(ctx context.Context, client *Client, args BdevLvolDecoupleParentArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "bdev_lvol_decouple_parent", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}
