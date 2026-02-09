package sdk

import (
	"fmt"
	"net/url"
)

// --- Threat Intelligence API (deprecated but functional) ---
// Endpoint: /services/data/threat_intel/item/{collection}

func threatIntelCollectionPath(collection, key string) string {
	base := fmt.Sprintf("/services/data/threat_intel/item/%s", URLEncode(collection))
	if key != "" {
		return fmt.Sprintf("%s/%s", base, URLEncode(key))
	}
	return base
}

// CreateThreatIntelItem creates an item in a threat intel collection.
func (c *SplunkClient) CreateThreatIntelItem(collection string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("ThreatIntel")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", threatIntelCollectionPath(collection, ""), params)
}

// ReadThreatIntelItem reads a threat intel item.
func (c *SplunkClient) ReadThreatIntelItem(collection, key string) (map[string]interface{}, error) {
	return c.DoRequest("GET", threatIntelCollectionPath(collection, key), nil)
}

// UpdateThreatIntelItem updates a threat intel item.
func (c *SplunkClient) UpdateThreatIntelItem(collection, key string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("ThreatIntel")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", threatIntelCollectionPath(collection, key), params)
}

// DeleteThreatIntelItem deletes a threat intel item.
func (c *SplunkClient) DeleteThreatIntelItem(collection, key string) error {
	lock := c.GetResourceLock("ThreatIntel")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(threatIntelCollectionPath(collection, key))
}

// ListThreatIntelItems lists items in a threat intel collection.
func (c *SplunkClient) ListThreatIntelItems(collection string) (map[string]interface{}, error) {
	return c.DoRequest("GET", threatIntelCollectionPath(collection, "")+"?count=-1", nil)
}

// UploadThreatIntel uploads a file (STIX/CSV/IOC) to threat intelligence.
func (c *SplunkClient) UploadThreatIntel(params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("ThreatIntel")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", "/services/data/threat_intel/upload", params)
}
