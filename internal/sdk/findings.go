package sdk

import (
	"fmt"
)

// --- ES v2 Findings API ---
// Base: /servicesNS/nobody/missioncontrol/public/v2/findings

const findingsBasePath = "/servicesNS/nobody/missioncontrol/public/v2/findings"

// CreateFinding creates a manual finding in ES.
func (c *SplunkClient) CreateFinding(body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("Finding")
	lock.Lock()
	defer lock.Unlock()

	return c.DoJSONRequest("POST", findingsBasePath, body)
}

// ReadFinding reads a finding by ID.
func (c *SplunkClient) ReadFinding(id string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/%s", findingsBasePath, URLEncode(id))
	return c.DoJSONRequest("GET", path, nil)
}

// ListFindings queries findings with optional parameters.
func (c *SplunkClient) ListFindings(filter string, limit int) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s?limit=%d", findingsBasePath, limit)
	if filter != "" {
		path += "&filter=" + URLEncode(filter)
	}
	return c.DoJSONRequest("GET", path, nil)
}
