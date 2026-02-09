package sdk

import (
	"fmt"
)

// --- ES v2 Investigations API ---
// Base: /servicesNS/nobody/missioncontrol/public/v2/investigations

const investigationsBasePath = "/servicesNS/nobody/missioncontrol/public/v2/investigations"

// CreateInvestigation creates a new ES investigation.
func (c *SplunkClient) CreateInvestigation(body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("Investigation")
	lock.Lock()
	defer lock.Unlock()

	return c.DoJSONRequest("POST", investigationsBasePath, body)
}

// ReadInvestigation retrieves investigations with optional query parameters.
func (c *SplunkClient) ReadInvestigation(id string) (map[string]interface{}, error) {
	path := investigationsBasePath + "?filter=" + URLEncode(fmt.Sprintf(`{"_key":"%s"}`, id))
	return c.DoJSONRequest("GET", path, nil)
}

// UpdateInvestigation updates an investigation by ID.
func (c *SplunkClient) UpdateInvestigation(id string, body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("Investigation")
	lock.Lock()
	defer lock.Unlock()

	path := fmt.Sprintf("%s/%s", investigationsBasePath, URLEncode(id))
	return c.DoJSONRequest("POST", path, body)
}

// ListInvestigations lists all investigations.
func (c *SplunkClient) ListInvestigations(limit int) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s?limit=%d", investigationsBasePath, limit)
	return c.DoJSONRequest("GET", path, nil)
}

// --- Investigation Notes ---

func investigationNotePath(investigationID, noteID string) string {
	base := fmt.Sprintf("%s/%s/notes", investigationsBasePath, URLEncode(investigationID))
	if noteID != "" {
		return fmt.Sprintf("%s/%s", base, URLEncode(noteID))
	}
	return base
}

// CreateInvestigationNote creates a note on an investigation.
func (c *SplunkClient) CreateInvestigationNote(investigationID string, body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("InvestigationNote")
	lock.Lock()
	defer lock.Unlock()

	return c.DoJSONRequest("POST", investigationNotePath(investigationID, ""), body)
}

// ReadInvestigationNotes reads notes for an investigation.
func (c *SplunkClient) ReadInvestigationNotes(investigationID string) (map[string]interface{}, error) {
	return c.DoJSONRequest("GET", investigationNotePath(investigationID, ""), nil)
}

// UpdateInvestigationNote updates a specific note.
func (c *SplunkClient) UpdateInvestigationNote(investigationID, noteID string, body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("InvestigationNote")
	lock.Lock()
	defer lock.Unlock()

	return c.DoJSONRequest("POST", investigationNotePath(investigationID, noteID), body)
}

// DeleteInvestigationNote deletes a specific note.
func (c *SplunkClient) DeleteInvestigationNote(investigationID, noteID string) error {
	lock := c.GetResourceLock("InvestigationNote")
	lock.Lock()
	defer lock.Unlock()

	_, err := c.DoJSONRequest("DELETE", investigationNotePath(investigationID, noteID), nil)
	return err
}

// --- Investigation Findings ---

// AddFindingsToInvestigation adds findings to an investigation.
func (c *SplunkClient) AddFindingsToInvestigation(investigationID string, body map[string]interface{}) (map[string]interface{}, error) {
	lock := c.GetResourceLock("Investigation")
	lock.Lock()
	defer lock.Unlock()

	path := fmt.Sprintf("%s/%s/findings", investigationsBasePath, URLEncode(investigationID))
	return c.DoJSONRequest("POST", path, body)
}
