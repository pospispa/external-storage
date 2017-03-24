package sharedfilesystems

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/volume"
)

// SharedFilesystemProvisioner is a class representing OpenStack Shared Filesystem external provisioner
type SharedFilesystemProvisioner struct {
	// Identity of this SharedFilesystemProvisioner, generated. Used to identify "this" provisioner's PVs.
	identity types.UID
}

// ZonesSCParamName is the name of the Storage Class parameter in which a set of zones is specified.
// The persistent volume will be dynamically provisioned in one of these zones.
const ZonesSCParamName = "zones"

const (
	// ProtocolNFS is the NFS shared filesystems protocol
	ProtocolNFS = "NFS"
)

func getPVCStorageSize(pvc *v1.PersistentVolumeClaim) (int, error) {
	errStorageSizeNotConfigured := fmt.Errorf("storage size request must be configured")
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
		if canonicalValue, noRounding := storageSize.AsScale(resource.Giga); !noRounding {
			return 0, fmt.Errorf("requested storage size must a be whole integer number in GBs")
		} else {
			var requiredButOmitted []byte
			storageSizeAsByte, _ := canonicalValue.AsCanonicalBytes(requiredButOmitted)
			if i, err := strconv.Atoi(string(storageSizeAsByte)); err != nil {
				return 0, fmt.Errorf("requested storage size is not an integer number")
			} else {
				return i, nil
			}
		}
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
	isZonesParam := false
	for index, value := range options.Parameters {
		switch strings.ToLower(index) {
		case ZonesSCParamName:
			if setOfZones, err := zonesToSet(value); err != nil {
				return request, err
			} else {
				request.AvailabilityZone = volume.ChooseZoneForVolume(setOfZones, options.PVC.Name)
				isZonesParam = true
			}
		default:
			return request, fmt.Errorf("invalid parameter %q", "foo")
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
