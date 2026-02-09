package sdk

import (
	"fmt"
)

// --- ES v2 Risks API ---
// Base: /servicesNS/nobody/missioncontrol/public/v2/risks/risk_scores

const risksBasePath = "/servicesNS/nobody/missioncontrol/public/v2/risks/risk_scores"

// ReadRiskScore reads the risk score for an entity.
func (c *SplunkClient) ReadRiskScore(entity, entityType string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/%s?entity_type=%s", risksBasePath, URLEncode(entity), URLEncode(entityType))
	return c.DoJSONRequest("GET", path, nil)
}

// AddRiskModifier adds a risk modifier for an entity.
func (c *SplunkClient) AddRiskModifier(entity string, body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("RiskModifier")
	lock.Lock()
	defer lock.Unlock()

	path := fmt.Sprintf("%s/%s", risksBasePath, URLEncode(entity))
	return c.DoJSONRequest("POST", path, body)
}
