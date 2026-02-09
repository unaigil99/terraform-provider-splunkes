package sdk

import (
	"fmt"
)

// --- ES v2 Assets API (read-only) ---
// Endpoint: /servicesNS/nobody/missioncontrol/public/v2/assets/{id}

const assetsBasePath = "/servicesNS/nobody/missioncontrol/public/v2/assets"

// ReadAsset reads an asset by ID.
func (c *SplunkClient) ReadAsset(id string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/%s", assetsBasePath, URLEncode(id))
	return c.DoJSONRequest("GET", path, nil)
}

// --- ES v2 Identity API (read-only) ---
// Endpoint: /servicesNS/nobody/missioncontrol/public/v2/identity/{id}

const identityBasePath = "/servicesNS/nobody/missioncontrol/public/v2/identity"

// ReadIdentity reads an identity by ID.
func (c *SplunkClient) ReadIdentity(id string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/%s", identityBasePath, URLEncode(id))
	return c.DoJSONRequest("GET", path, nil)
}
