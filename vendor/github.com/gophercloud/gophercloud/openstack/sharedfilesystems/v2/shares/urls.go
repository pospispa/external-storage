package shares

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud"
)

func createURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("shares")
}

func deleteURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("shares", id)
}

func getURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("shares", id)
}

func getMicroversionsURL(c *gophercloud.ServiceClient) string {
	baseURLWithoutEndingSlashes := strings.TrimRight(c.ResourceBaseURL(), "/")
	slashIndexBeforeProjectID := strings.LastIndex(baseURLWithoutEndingSlashes, "/")
	slashIndexBeforeProjectID = strings.LastIndex(baseURLWithoutEndingSlashes[:slashIndexBeforeProjectID], "/")
	fmt.Println("")
	fmt.Printf("url: %q", baseURLWithoutEndingSlashes[:slashIndexBeforeProjectID]+"/")
	fmt.Println("")
	return baseURLWithoutEndingSlashes[:slashIndexBeforeProjectID] + "/"
}

func grantAccessURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("shares", id, "action")
}

func getExportLocationsURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("shares", id, "export_locations")
}
