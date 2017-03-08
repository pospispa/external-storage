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

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
	"k8s.io/kubernetes/pkg/volume"
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
	errStorageSizeNotConfigured := fmt.Errorf("requested storage capacity must be set")
	if pvc.Spec.Resources.Requests == nil {
		return 0, errStorageSizeNotConfigured
	}
	if storageSize, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]; !ok {
		return 0, errStorageSizeNotConfigured
	} else {
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
}

func PrepareCreateRequest(options controller.VolumeOptions, getAllZones func() (sets.String, error)) (shares.CreateOpts, error) {
	var request shares.CreateOpts
	// Currently only the NFS shares are supported, that's why the NFS is hardcoded.
	request.ShareProto = ProtocolNFS
	// mandatory parameters
	if storageSize, err := getPVCStorageSize(options.PVC); err != nil {
		return request, err
	} else {
		request.Size = storageSize
	}

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
			if setOfZones, err := volume.ZonesToSet(value); err != nil {
				return request, err
			} else {
				request.AvailabilityZone = volume.ChooseZoneForVolume(setOfZones, options.PVC.Name)
				isZonesParam = true
			}
		case TypeSCParamName:
			request.ShareType = value
		default:
			return request, fmt.Errorf("invalid parameter %q", index)
		}
	}
	if !isZonesParam {
		if allAvailableZones, err := getAllZones(); err != nil {
			return request, err
		} else {
			request.AvailabilityZone = volume.ChooseZoneForVolume(allAvailableZones, options.PVC.Name)
		}
	}
	return request, nil
}
