package sharedfilesystems

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/util/sets"
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
