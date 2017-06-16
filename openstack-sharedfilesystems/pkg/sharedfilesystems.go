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

package sharedfilesystems

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
)

const (
	// ZonesSCParamName is the name of the Storage Class parameter in which a set of zones is specified.
	// The persistent volume will be dynamically provisioned in one of these zones.
	ZonesSCParamName = "zones"
	// TypeSCParamName is the name of a share type configured by administrator of Manila service.
	TypeSCParamName = "type"
	// ProtocolNFS is the NFS shared filesystems protocol
	ProtocolNFS = "NFS"
)

func getPVCStorageSize(pvc *v1.PersistentVolumeClaim) (int, error) {
	var storageSize resource.Quantity
	var ok bool
	errStorageSizeNotConfigured := fmt.Errorf("requested storage capacity must be set")
	if pvc.Spec.Resources.Requests == nil {
		return 0, errStorageSizeNotConfigured
	}
	if storageSize, ok = pvc.Spec.Resources.Requests[v1.ResourceStorage]; !ok {
		return 0, errStorageSizeNotConfigured
	}
	if storageSize.IsZero() {
		return 0, fmt.Errorf("requested storage size must not have zero value")
	}
	if storageSize.Sign() == -1 {
		return 0, fmt.Errorf("requested storage size must be greater than zero")
	}
	canonicalValue, _ := storageSize.AsScale(resource.Giga)
	var buf []byte
	storageSizeAsByte, _ := canonicalValue.AsCanonicalBytes(buf)
	var i int
	var err error
	if i, err = strconv.Atoi(string(storageSizeAsByte)); err != nil {
		return 0, fmt.Errorf("requested storage size is not a number")
	}
	return i, nil
}

// PrepareCreateRequest return:
// - success: ready to send shared filesystem create request data structure constructed from Persistent Volume Claim and corresponding Storage Class
// - failure: an error
func PrepareCreateRequest(options controller.VolumeOptions, getAllZones func() (sets.String, error)) (shares.CreateOpts, error) {
	var request shares.CreateOpts
	var storageSize int
	var err error
	// Currently only the NFS shares are supported, that's why the NFS is hardcoded.
	request.ShareProto = ProtocolNFS
	// mandatory parameters
	if storageSize, err = getPVCStorageSize(options.PVC); err != nil {
		return request, err
	}
	request.Size = storageSize

	// optional parameters
	request.Name = "pvc-" + string(options.PVC.UID)
	tags := make(map[string]string)
	tags[persistentvolume.CloudVolumeCreatedForClaimNamespaceTag] = options.PVC.Namespace
	tags[persistentvolume.CloudVolumeCreatedForClaimNameTag] = options.PVC.Name
	tags[persistentvolume.CloudVolumeCreatedForVolumeNameTag] = request.Name
	request.Metadata = tags
	isZonesParam := false
	for index, value := range options.Parameters {
		switch strings.ToLower(index) {
		case ZonesSCParamName:
			setOfZones, err := util.ZonesToSet(value)
			if err != nil {
				return request, err
			}
			request.AvailabilityZone = volume.ChooseZoneForVolume(setOfZones, options.PVC.Name)
			isZonesParam = true
		case TypeSCParamName:
			request.ShareType = value
		default:
			return request, fmt.Errorf("invalid parameter %q", index)
		}
	}
	if !isZonesParam {
		var allAvailableZones sets.String
		var err error
		if allAvailableZones, err = getAllZones(); err != nil {
			return request, err
		}
		request.AvailabilityZone = volume.ChooseZoneForVolume(allAvailableZones, options.PVC.Name)
	}
	return request, nil
}

// WaitTillAvailable keeps querying Manila API for a share status until it is available. The waiting can:
// - succeed: in this case the is/becomes available
// - timeout: error is returned.
// - another error occurs: error is returned.
func WaitTillAvailable(client *gophercloud.ServiceClient, shareID string) error {
	desiredStatus := "available"
	timeout := 120 /* secs */
	return gophercloud.WaitFor(timeout, func() (bool, error) {
		current, err := shares.Get(client, shareID).Extract()
		if err != nil {
			return false, err
		}

		if current.Status == desiredStatus {
			return true, nil
		}
		return false, nil
	})
}

// ChooseExportLocation chooses one ExportLocation according to the below rules:
// 1. Path is not empty, i.e. is not an empty string or does not contain spaces and tabs only
// 2. IsAdminOnly == false
// 3. Preferred == true are preferred over Preferred == false
// 4. Locations with lower slice index are preferred over locations with higher slice index
// In case no location complies with the above rules an error is returned.
func ChooseExportLocation(locs []shares.ExportLocation) (shares.ExportLocation, error) {
	if len(locs) == 0 {
		return shares.ExportLocation{}, fmt.Errorf("Error: received an empty list of export locations")
	}
	foundMatchingNotPreferred := false
	var matchingNotPreferred shares.ExportLocation
	for _, loc := range locs {
		if loc.IsAdminOnly || strings.TrimSpace(loc.Path) == "" {
			continue
		}
		if loc.Preferred {
			return loc, nil
		}
		if !foundMatchingNotPreferred {
			matchingNotPreferred = loc
			foundMatchingNotPreferred = true
		}
	}
	if foundMatchingNotPreferred {
		return matchingNotPreferred, nil
	}
	return shares.ExportLocation{}, fmt.Errorf("cannot find any non-admin export location")
}
