/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	sharedfilesystems "github.com/kubernetes-incubator/external-storage/openstack-sharedfilesystems/pkg"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kubernetes/pkg/util/sets"
)

func devMockGetAllZones() (sets.String, error) {
	ret := sets.String{"nova": sets.Empty{}}
	return ret, nil
}

func main() {
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		fmt.Printf("AuthOptionsFromEnv failed: (%v)", err)
		return
	}
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		fmt.Printf("AuthenticatedClient failed: (%v)", err)
		return
	}
	client, err := openstack.NewSharedFileSystemV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	if err != nil {
		fmt.Printf("NewSharedFileSystemV2 failed: (%v)", err)
		return
	}

	pvc := controller.VolumeOptions{
		PersistentVolumeReclaimPolicy: "Delete",
		PVName: "pv",
		PVC: &v1.PersistentVolumeClaim{
			ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
			Spec: v1.PersistentVolumeClaimSpec{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.Quantity{},
					},
				},
			},
		},
		Parameters: map[string]string{sharedfilesystems.ZonesSCParamName: "nova"},
	}
	storageSize := "2G"
	if quantity, err := resource.ParseQuantity(storageSize); err != nil {
		fmt.Printf("Failed to parse storage size (%v): %v", storageSize, err)
	} else {
		pvc.PVC.Spec.Resources.Requests[v1.ResourceStorage] = quantity
	}
	if createReq, err := sharedfilesystems.PrepareCreateRequest(pvc, devMockGetAllZones); err != nil {
		fmt.Printf("Failed to create Create Request: %v", err)
	} else {
		if createReqResponse, err := shares.Create(client, createReq).Extract(); err != nil {
			fmt.Printf("Response to create request says failed: %v", err)
		} else {
			fmt.Printf("Response: (%v)", createReqResponse)
		}
	}
}
