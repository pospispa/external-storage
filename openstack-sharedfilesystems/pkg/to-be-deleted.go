package sharedfilesystems

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/util/sets"
)

// TO BE DELETED after the below code is merged into K8s

// CloudVolumeCreatedForClaimNamespaceTag is a name of a tag attached to a real volume in cloud (e.g. AWS EBS or GCE PD)
// with namespace of a persistent volume claim used to create this volume.
const CloudVolumeCreatedForClaimNamespaceTag = "kubernetes.io/created-for/pvc/namespace"

// CloudVolumeCreatedForClaimNameTag is a name of a tag attached to a real volume in cloud (e.g. AWS EBS or GCE PD)
// with name of a persistent volume claim used to create this volume.
const CloudVolumeCreatedForClaimNameTag = "kubernetes.io/created-for/pvc/name"

// CloudVolumeCreatedForVolumeNameTag is a name of a tag attached to a real volume in cloud (e.g. AWS EBS or GCE PD)
// with name of appropriate Kubernetes persistent volume .
const CloudVolumeCreatedForVolumeNameTag = "kubernetes.io/created-for/pv/name"

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
