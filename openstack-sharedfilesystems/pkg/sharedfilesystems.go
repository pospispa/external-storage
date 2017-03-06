package sharedfilesystems

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
)

// TO BE DELETED after the below function(s) are merged into k8s

// zonesToSet converts a string containing a comma separated list of zones to set
func zonesToSet(zonesString string) (sets.String, error) {
	zonesSlice := strings.Split(zonesString, ",")
	zonesSet := make(sets.String)
	for _, zone := range zonesSlice {
		trimmedZone := strings.TrimSpace(zone)
		if trimmedZone == "" {
			return make(sets.String), fmt.Errorf("comma separated list of zones (%q) must not contain an empty zone", zonesString)
		}
		zonesSet.Insert(trimmedZone)
	}
	return zonesSet, nil
}

// TO BE DELETED after the above function(s) are merged into k8s

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

func prepareCreateRequest(options controller.VolumeOptions) (shares.CreateOpts, error) {
	var request shares.CreateOpts
	// Currently only the NFS shares are supported, that's why the NFS is hardcoded.
	request.ShareProto = ProtocolNFS
	// mandatory parameters
	if storageSize, err := getPVCStorageSize(options.PVC); err != nil {
		return request, err
	} else {
		request.Size = storageSize
	}

	// optional parameter
	for index, _ := range options.Parameters {
		switch strings.ToLower(index) {
		case ZonesSCParamName:
		default:
			return request, fmt.Errorf("invalid parameter %q", "foo")
		}
	}
	return request, nil
}
