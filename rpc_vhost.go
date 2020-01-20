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

type VhostCreateBlkControllerArgs struct {
	Ctrlr    string `json:"ctrlr"`
	DevName  string `json:"dev_name"`
	Readonly bool   `json:"readonly,omitempty"`
	Cpumask  string `json:"cpumask,omitempty"`
}

//VhostCreateBlkControllerResponse is bool: indication of result
func VhostCreateBlkController(ctx context.Context, client *Client, args VhostCreateBlkControllerArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "vhost_create_blk_controller", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}

type VhostDeleteControllerArgs struct {
	Ctrlr string `json:"ctrlr"`
}

//VhostDeleteControllerResponse is bool: indication of result
func VhostDeleteController(ctx context.Context, client *Client, args VhostDeleteControllerArgs) (bool, error) {
	var response bool
	err := client.Invoke(ctx, "vhost_delete_controller", args, &response)
	if err != nil {
		return false, err
	}
	return response, err
}

type VhostGetControllersResponse []Controller

type Controller struct {
	Ctrlr         string `json:"ctrlr"`
	Cpumask       string `json:"cpumask"`
	DelayBaseUs   int    `json:"delay_base_us"`
	IposThreshold int    `json:"iops_threshold"`
	// BackendSpecific holds the parsed JSON response for known
	// backends (like SCSIControllerSpecific), otherwise
	// the JSON data converted to basic types (map, list, etc.)
	BackendSpecific BackendSpecificType `json:"backend_specific"`
}

type BackendSpecificType map[string]interface{}

type VhostScsiBackendSpecific []VhostScsiBackend
type VhostNvmeBackendSpecific []VhostNvmeBackend
type VhostBlkBackendSpecific struct {
	Bdev     string `json:"bdev"`
	Readonly bool   `json: readonly`
}

type VhostNvmeBackend struct {
	Nsid int32  `json:"nsid"`
	Bdev string `json:"bdev"`
}

type VhostScsiBackend struct {
	TargetName string
	ID         int32
	ScsiDevNum uint32
	Luns       []VhostScsiLun
}

type VhostScsiLun struct {
	ID       int32  `json:"id"`
	BdevName string `json:"bdev_name"`
}

func getBlkBackendSpecific(in interface{}) VhostBlkBackendSpecific {
	target := VhostBlkBackendSpecific{}
	if hash, ok := in.(map[string]interface{}); ok {
		for key, value := range hash {
			switch key {
			case "bdev":
				if name, ok := value.(string); ok {
					target.Bdev = name
				}
			case "readonly":
				if readonly, ok := value.(bool); ok {
					target.Readonly = readonly
				}
			}
		}
	}

	return target
}

func getNvmeBackendSpecific(in interface{}) VhostNvmeBackendSpecific {
	result := VhostNvmeBackendSpecific{}
	list, ok := in.([]interface{})
	if !ok {
		return result
	}
	for _, entry := range list {
		if hash, ok := entry.(map[string]interface{}); ok {
			target := VhostNvmeBackend{}
			for key, value := range hash {
				switch key {
				case "bdev":
					if name, ok := value.(string); ok {
						target.Bdev = name
					}
				case "nsid":
					if nsid, ok := value.(float64); ok {
						target.Nsid = int32(nsid)
					}
				}
			}
			result = append(result, target)
		}
	}

	return result
}

func getScsiBackendLuns(ins []interface{}) []VhostScsiLun {
	luns := []VhostScsiLun{}

	for _, lun := range ins {
		var l VhostScsiLun
		if hash, ok := lun.(map[string]interface{}); ok {
			for key, value := range hash {
				switch key {
				case "id":
					if id, ok := value.(float64); ok {
						l.ID = int32(id)
					}
				case "bdev_name":
					if name, ok := value.(string); ok {
						l.BdevName = name
					}
				}
			}
		}
		luns = append(luns, l)
	}

	return luns
}

// getSCSIBackendSpecific interprets the Controller.BackendSpecific value for
// map entries with key "scsi". See https://github.com/spdk/spdk/issues/329#issuecomment-396266197
// and spdk_vhost_scsi_dump_info_json().
func getScsiBackendSpecific(in interface{}) VhostScsiBackendSpecific {
	result := VhostScsiBackendSpecific{}
	list, ok := in.([]interface{})
	if !ok {
		return result
	}
	for _, entry := range list {
		if hash, ok := entry.(map[string]interface{}); ok {
			target := VhostScsiBackend{
				Luns: []VhostScsiLun{},
			}
			for key, value := range hash {
				switch key {
				case "target_name":
					if name, ok := value.(string); ok {
						target.TargetName = name
					}
				case "id":
					if id, ok := value.(float64); ok {
						target.ID = int32(id)
					}
				case "scsi_dev_num":
					if devNum, ok := value.(float64); ok {
						target.ScsiDevNum = uint32(devNum)
					}
				case "luns":
					if luns, ok := value.([]interface{}); ok {
						target.Luns = getScsiBackendLuns(luns)
					}
				}
			}
			result = append(result, target)
		}
	}
	return result
}

type VhostGetControllersArgs struct {
	Name string `json:"name,omitempty"`
}

func VhostGetControllers(ctx context.Context, client *Client, args VhostGetControllersArgs) (VhostGetControllersResponse, error) {
	var response VhostGetControllersResponse
	err := client.Invoke(ctx, "vhost_get_controllers", args, &response)
	if err == nil {
		for _, controller := range response {
			for backend, specific := range controller.BackendSpecific {
				switch backend {
				case "block":
					controller.BackendSpecific[backend] = getBlkBackendSpecific(specific)

				case "scsi":
					controller.BackendSpecific[backend] = getScsiBackendSpecific(specific)

				case "namespaces":
					controller.BackendSpecific[backend] = getNvmeBackendSpecific(specific)
				}
			}
		}
	}
	return response, err
}
