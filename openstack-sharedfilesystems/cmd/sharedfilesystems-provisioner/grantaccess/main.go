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
		fmt.Println("")
		return
	}
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		fmt.Printf("AuthenticatedClient failed: (%v)", err)
		fmt.Println("")
		return
	}
	client, err := openstack.NewSharedFileSystemV2(provider, gophercloud.EndpointOpts{Region: "RegionOne"})
	if err != nil {
		fmt.Printf("NewSharedFileSystemV2 failed: (%v)", err)
		fmt.Println("")
		return
	}
	fmt.Printf("Authorization options: (%v)", authOpts)
	fmt.Println("")
	fmt.Printf("Provider client: (%v)", provider)
	fmt.Println("")
	fmt.Printf("Client: (%v)", client)
	fmt.Println("")
	fmt.Printf("Client endpoint: (%v)", client.Endpoint)
	fmt.Println("")
	fmt.Printf("Client resource base: (%v)", client.ResourceBase)
	fmt.Println("")
	fmt.Printf("Client microversion: (%v)", client.Microversion)
	fmt.Println("")

	//shareID := "f245062e-06c1-410c-a0ef-20fa036a1071"
	// fda
	//	var grantAccessReq shares.GrantAccessOpts
	//	grantAccessReq.AccessType = "ip"
	//	grantAccessReq.AccessTo = "0.0.0.0/0"
	//	grantAccessReq.AccessLevel = "rw"
	// var tenantID string
	//grantAccessReq.TenantID = tenantID
	//	grantAccessReqResponse := shares.GrantAccess(client, grantAccessReq, shareID)
	//	fmt.Printf("Grant Access response before extract: (%v)", grantAccessReqResponse)
	//	fmt.Println("")
	//	if extractedGrantAccessReqResponse, err := grantAccessReqResponse.ExtractGrantAccess(); err != nil {
	//		fmt.Printf("Response to grant access request says failed: (%v)", err)
	//		fmt.Println("")
	//		return
	//	} else {
	//		fmt.Printf("Grant Access response after extract: (%v)", extractedGrantAccessReqResponse)
	//		fmt.Println("")
	//	}

	//const (
	//	v20 = "v2.0"
	//	v30 = "v3.0"
	//)

	//versions := []*gc_utils.Version{
	//	{ID: v20, Priority: 20, Suffix: "/v2.0/"},
	//	{ID: v30, Priority: 30, Suffix: "/v3/"},
	//}
	//if version, endpoint, err := gc_utils.ChooseVersion(provider, versions); err != nil {
	//	fmt.Printf("ChooseVersion returned error: (%v)", err)
	//	fmt.Println("")
	//} else {
	//	fmt.Printf("ChooseVersion, version: (%v)", version)
	//	fmt.Println("")
	//	fmt.Printf("ChooseVersion, endpoint: (%v)", endpoint)
	//	fmt.Println("")
	//}

	respMicroVersions := shares.GetMicroversion(client)
	fmt.Printf("Microversions response: (%v)", respMicroVersions)
	fmt.Println("")
	if extractedMicroVersionsReqResp, err := respMicroVersions.ExtractMicroversion(); err != nil {
		fmt.Printf("Extraction of Microversions Response failed: (%v)", err)
		fmt.Println("")
		return
	} else {
		fmt.Printf("Microversions response after extraction: (%v)", extractedMicroVersionsReqResp)
		fmt.Println("")
		fmt.Printf("Microversion status: (%v)", (*extractedMicroVersionsReqResp)[0].Status)
		fmt.Println("")
		fmt.Printf("Microversion: (%v)", (*extractedMicroVersionsReqResp)[0].Version)
		fmt.Println("")
	}
}
