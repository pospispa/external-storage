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
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
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
	flag.Parse()
	flag.Set("logtostderr", "true")

	regionName := os.Getenv("OS_REGION_NAME")
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		glog.Fatalf("%v", err)
	}
	glog.V(1).Infof("successfully read options from environment variables: (%v)", authOpts)
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		glog.Fatalf("authentication failed: (%v)", err)
	}
	glog.V(4).Infof("successfully created provider client: (%v)", provider)
	client, err := openstack.NewSharedFileSystemV2(provider, gophercloud.EndpointOpts{Region: regionName})
	if err != nil {
		glog.Fatalf("failed to create Manila v2 client: (%v)", err)
	}
	client.Microversion = "2.21"
	serverVer, err := apiversions.Get(client, "v2").Extract()
	if err != nil {
		glog.Fatalf("failed to get Manila v2 API min/max microversions: (%v)", err)
	}
	glog.V(4).Infof("received server's microvesion data structure: (%v)", serverVer)
	glog.V(3).Infof("server's min microversion is: %q, max microversion is: %q", serverVer.MinVersion, serverVer.Version)
	if err = sharedfilesystems.ValidMicroversion(serverVer.MinVersion); err != nil {
		glog.Fatalf("server's minimum microversion is invalid: (%v)", serverVer.MinVersion)
	}
	if err = sharedfilesystems.ValidMicroversion(serverVer.Version); err != nil {
		glog.Fatalf("server's maximum microversion is invalid: (%v)", serverVer.Version)
	}
	clientMajor, clientMinor := sharedfilesystems.SplitMicroversion(client.Microversion)
	minMajor, minMinor := sharedfilesystems.SplitMicroversion(serverVer.MinVersion)
	if clientMajor < minMajor || (clientMajor == minMajor && clientMinor < minMinor) {
		glog.Fatalf("client microversion (%q) is smaller than the server's minimum microversion (%q)", client.Microversion, serverVer.MinVersion)
	}
	maxMajor, maxMinor := sharedfilesystems.SplitMicroversion(serverVer.Version)
	if maxMajor < clientMajor || (maxMajor == clientMajor && maxMinor < clientMinor) {
		glog.Fatalf("client microversion (%q) is bigger than the server's maximum microversion (%q)", client.Microversion, serverVer.Version)
	}
	glog.V(4).Infof("successfully created Manila v2 client: (%v)", client)

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
						v1.ResourceStorage: resource.MustParse("2G"),
					},
				},
			},
		},
		// Manila on notebook
		Parameters: map[string]string{sharedfilesystems.ZonesSCParamName: "nova"},
		// Roger's OpenStack
		// Parameters: map[string]string{sharedfilesystems.ZonesSCParamName: "nova", TypeSCParamName: "default"},
	}
	var createdShare shares.Share
	var createReq shares.CreateOpts
	if createReq, err = sharedfilesystems.PrepareCreateRequest(pvc, devMockGetAllZones); err != nil {
		glog.Errorf("failed to create Create Request: (%v)", err)
		return
	}
	glog.V(4).Infof("successfully created a share Create Request: %v", createReq)
	var createReqResponse *shares.Share
	if createReqResponse, err = shares.Create(client, createReq).Extract(); err != nil {
		glog.Errorf("failed to create a share: (%v)", err)
		return
	}
	glog.V(3).Infof("successfully created a share: (%v)", createReqResponse)
	createdShare = *createReqResponse
	if err = sharedfilesystems.WaitTillAvailable(client, createdShare.ID); err != nil {
		glog.Errorf("waiting for the share %q to become created failed: (%v)", createdShare.ID, err)
		return
	}
	glog.V(4).Infof("the share %q is now in state created", createdShare.ID)

	var grantAccessReq shares.GrantAccessOpts
	grantAccessReq.AccessType = "ip"
	grantAccessReq.AccessTo = "0.0.0.0/0"
	grantAccessReq.AccessLevel = "rw"
	var grantAccessReqResponse *shares.AccessRight
	if grantAccessReqResponse, err = shares.GrantAccess(client, createdShare.ID, grantAccessReq).Extract(); err != nil {
		glog.Errorf("failed to grant access to the share %q: (%v)", createdShare.ID, err)
		return
	}
	glog.V(4).Infof("granted access to the share %q: (%v)", createdShare.ID, grantAccessReqResponse)

	var exportLocations []shares.ExportLocation
	var chosenLocation shares.ExportLocation
	var getExportLocationsReqResponse []shares.ExportLocation
	if getExportLocationsReqResponse, err = shares.GetExportLocations(client, createdShare.ID).Extract(); err != nil {
		glog.Errorf("failed to get export locations for the share %q: (%v)", createdShare.ID, err)
		return
	}
	glog.V(4).Infof("got export locations for the share %q: (%v)", createdShare.ID, getExportLocationsReqResponse)
	exportLocations = getExportLocationsReqResponse
	if chosenLocation, err = sharedfilesystems.ChooseExportLocation(exportLocations); err != nil {
		fmt.Printf("failed to choose an export location for the share %q: %q", createdShare.ID, err.Error())
		return
	}
	glog.V(4).Infof("selected export location for the share %q is: (%v)", createdShare.ID, chosenLocation)
	pv, err := sharedfilesystems.FillInPV(pvc, createdShare, chosenLocation)
	if err != nil {
		glog.Errorf("failed to fill in PV for the share %q: %q", createdShare.ID, err.Error())
		return
	}
	glog.V(4).Infof("resulting PV for the share %q: (%v)", createdShare.ID, pv)
}
