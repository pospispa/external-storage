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
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/apiversions"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	sharedfilesystems "github.com/kubernetes-incubator/external-storage/openstack-sharedfilesystems/pkg"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func devMockGetAllZones() (sets.String, error) {
	ret := sets.String{"nova": sets.Empty{}}
	return ret, nil
}

func main() {
	regionName := os.Getenv("OS_REGION_NAME")
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		fmt.Printf("AuthOptionsFromEnv failed: (%v)", err)
		fmt.Println("")
		return
	}
	fmt.Println("")
	fmt.Printf("AuthOptionsFromEnv: (%v)", authOpts)
	fmt.Println("")
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		fmt.Printf("AuthenticatedClient failed: (%v)", err)
		fmt.Println("")
		return
	}
	fmt.Println("")
	fmt.Printf("Provider client: (%v)", provider)
	fmt.Println("")
	client, err := openstack.NewSharedFileSystemV2(provider, gophercloud.EndpointOpts{Region: regionName})
	if err != nil {
		fmt.Printf("NewSharedFileSystemV2 failed: (%v)", err)
		fmt.Println("")
		return
	}
	client.Microversion = "2.21"
	serverVer, err := apiversions.Get(client, "v2").Extract()
	if err != nil {
		fmt.Printf("apiversions.Get failed: (%v)", err)
		fmt.Println("")
		return
	} else {
		fmt.Printf("apiversions.Get returned: (%v)", serverVer)
		fmt.Printf("Server's min microversion: %q, max microversion: %q", serverVer.MinVersion, serverVer.Version)
		fmt.Println("")
	}
	if err = sharedfilesystems.ValidMicroversion(serverVer.MinVersion); err != nil {
		fmt.Printf("apiversions.Get returned invalid minimum microversion: (%v)", serverVer.MinVersion)
		fmt.Println("")
		return
	}
	if err = sharedfilesystems.ValidMicroversion(serverVer.Version); err != nil {
		fmt.Printf("apiversions.Get returned invalid maximum microversion: (%v)", serverVer.Version)
		fmt.Println("")
		return
	}
	clientMajor, clientMinor := sharedfilesystems.SplitMicroversion(client.Microversion)
	minMajor, minMinor := sharedfilesystems.SplitMicroversion(serverVer.MinVersion)
	if clientMajor < minMajor || (clientMajor == minMajor && clientMinor < minMinor) {
		fmt.Printf("client microversion (%q) is smaller than server's min microversion (%q)", client.Microversion, serverVer.MinVersion)
		fmt.Println("")
		return
	}
	maxMajor, maxMinor := sharedfilesystems.SplitMicroversion(serverVer.Version)
	if maxMajor < clientMajor || (maxMajor == clientMajor && maxMinor < clientMinor) {
		fmt.Printf("client microversion (%q) is bigger than server's max microversion (%q)", client.Microversion, serverVer.Version)
		fmt.Println("")
		return
	}
	fmt.Printf("Service client: (%v)", client)
	fmt.Println("")

	pvc := controller.VolumeOptions{
		PersistentVolumeReclaimPolicy: "Delete",
		PVName: "pv",
		PVC: &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pvc",
				Namespace: "foo",
				UID:       types.UID("unique-uid")},
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
	var createdShare shares.Share
	if createReq, err := sharedfilesystems.PrepareCreateRequest(pvc, devMockGetAllZones); err != nil {
		fmt.Printf("Failed to create Create Request: (%v)", err)
	} else {
		fmt.Printf("Request: %v", createReq)
		fmt.Println("")
		if createReqResponse, err := shares.Create(client, createReq).Extract(); err != nil {
			fmt.Printf("Response to create request says failed: (%v)", err)
			fmt.Println("")
			return
		} else {
			fmt.Printf("Create response: (%v)", createReqResponse)
			fmt.Println("")
			createdShare = *createReqResponse
		}
	}
	fmt.Println("")
	if err = sharedfilesystems.WaitTillAvailable(client, createdShare.ID); err != nil {
		fmt.Printf("Response to WaitTillAvailable says failed: (%v)", err)
		fmt.Println("")
		return
	} else {
		fmt.Printf("WaitTillAvailable returned no error")
		fmt.Println("")
	}

	var grantAccessReq shares.GrantAccessOpts
	grantAccessReq.AccessType = "ip"
	grantAccessReq.AccessTo = "0.0.0.0/0"
	grantAccessReq.AccessLevel = "rw"
	if grantAccessReqResponse, err := shares.GrantAccess(client, createdShare.ID, grantAccessReq).Extract(); err != nil {
		fmt.Printf("Response to grant access request says failed: (%v)", err)
		fmt.Println("")
		return
	} else {
		fmt.Printf("Grant Access response: (%v)", grantAccessReqResponse)
		fmt.Println("")
	}

	fmt.Println("")
	var exportLocations []shares.ExportLocation
	var chosenLocation shares.ExportLocation
	if getExportLocationsReqResponse, err := shares.GetExportLocations(client, createdShare.ID).Extract(); err != nil {
		fmt.Printf("Response to get export locations request says failed: (%v)", err)
		fmt.Println("")
		return
	} else {
		fmt.Printf("Get Export Locations response: (%v)", getExportLocationsReqResponse)
		fmt.Println("")
		exportLocations = getExportLocationsReqResponse
	}
	if chosenLocation, err = sharedfilesystems.ChooseExportLocation(exportLocations); err != nil {
		fmt.Println("")
		fmt.Printf("Failed to choose an export location: %q", err.Error())
		fmt.Println("")
	} else {
		fmt.Println("")
		fmt.Printf("chosen export location: (%v)", chosenLocation)
		fmt.Println("")
	}
	pv, err := sharedfilesystems.FillInPV(pvc, createdShare, chosenLocation)
	if err != nil {
		fmt.Println("")
		fmt.Printf("Failed to fill in PV: %q", err.Error())
		fmt.Println("")
	} else {
		fmt.Println("")
		fmt.Printf("Resulting PV: (%v)", pv)
		fmt.Println("")
	}
}
