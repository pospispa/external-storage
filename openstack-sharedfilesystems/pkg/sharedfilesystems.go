package sharedfilesystems

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
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

// PrepareCreateRequest return:
// - success: ready to send shared filesystem create request data structure constructed from Persistent Volume Claim and corresponding Storage Class
// - failure: an error
func PrepareCreateRequest(options controller.VolumeOptions, getAllZones func() (sets.String, error)) (shares.CreateOpts, error) {
	var request shares.CreateOpts
	// Currently only the NFS shares are supported, that's why the NFS is hardcoded.
	// Manila on notebook
	request.ShareProto = ProtocolNFS
	// Roger's OpenStack
	//request.ShareProto = "CEPHFS"
	//request.ShareType = "default"

	// mandatory parameters
	if storageSize, err := getPVCStorageSize(options.PVC); err != nil {
		return request, err
	} else {
		request.Size = storageSize
	}

	// optional parameters
	request.Name = "pvc-" + string(options.PVC.UID)
	tags := make(map[string]string)
	tags[CloudVolumeCreatedForClaimNamespaceTag] = options.PVC.Namespace
	tags[CloudVolumeCreatedForClaimNameTag] = options.PVC.Name
	tags[CloudVolumeCreatedForVolumeNameTag] = request.Name
	request.Metadata = tags
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

// WaitTillAvailable keeps querying Manila API for a share status until it is available. The waiting can:
// - succeed: in this case the Get response is returned.
// - timeout: error is returned.
// - another error occurs: error is returned.
func WaitTillAvailable(client *gophercloud.ServiceClient, shareID string) (*shares.Share, error) {
	desiredState := "available"
	var timeoutInSec, firstWaitInSec, waitMultiplier, currentWaitInSec time.Duration
	timeoutInSec = 120 * 1000 * time.Millisecond
	firstWaitInSec = 1 * 1000 * time.Millisecond
	waitMultiplier = 2
	currentWaitInSec = firstWaitInSec
	for currentWaitInSec <= timeoutInSec {
		time.Sleep(currentWaitInSec)
		if getReqResponse, err := shares.Get(client, shareID).Extract(); err != nil {
			return nil, err
		} else {
			if getReqResponse.Status == desiredState {
				return getReqResponse, nil
			}
		}
		currentWaitInSec *= waitMultiplier
	}
	return nil, fmt.Errorf("timeouted waiting for the provisioned share (id: %q) to become available.", shareID)
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
	return shares.ExportLocation{}, fmt.Errorf("Error: not found any non-AdminOnly export location")
}
