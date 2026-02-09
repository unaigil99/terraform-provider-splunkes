package sdk

import (
	"fmt"
	"net/url"
)

// --- Lookup Definitions CRUD ---
// Endpoint: /servicesNS/{owner}/{app}/data/transforms/lookups

func lookupDefinitionPath(owner, app, name string) string {
	if name != "" {
		return fmt.Sprintf("/servicesNS/%s/%s/data/transforms/lookups/%s", URLEncode(owner), URLEncode(app), URLEncode(name))
	}
	return fmt.Sprintf("/servicesNS/%s/%s/data/transforms/lookups", URLEncode(owner), URLEncode(app))
}

// CreateLookupDefinition creates a new lookup definition.
func (c *SplunkClient) CreateLookupDefinition(owner, app string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("LookupDefinition")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", lookupDefinitionPath(owner, app, ""), params)
}

// ReadLookupDefinition reads a lookup definition by name.
func (c *SplunkClient) ReadLookupDefinition(owner, app, name string) (map[string]interface{}, error) {
	return c.DoRequest("GET", lookupDefinitionPath(owner, app, name), nil)
}

// UpdateLookupDefinition updates a lookup definition.
func (c *SplunkClient) UpdateLookupDefinition(owner, app, name string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("LookupDefinition")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", lookupDefinitionPath(owner, app, name), params)
}

// DeleteLookupDefinition deletes a lookup definition.
func (c *SplunkClient) DeleteLookupDefinition(owner, app, name string) error {
	lock := c.GetResourceLock("LookupDefinition")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(lookupDefinitionPath(owner, app, name))
}

// --- Lookup Table Files CRUD ---
// Endpoint: /servicesNS/{owner}/{app}/data/lookup-table-files

func lookupTablePath(owner, app, name string) string {
	if name != "" {
		return fmt.Sprintf("/servicesNS/%s/%s/data/lookup-table-files/%s", URLEncode(owner), URLEncode(app), URLEncode(name))
	}
	return fmt.Sprintf("/servicesNS/%s/%s/data/lookup-table-files", URLEncode(owner), URLEncode(app))
}

// CreateLookupTable creates a new lookup table file.
func (c *SplunkClient) CreateLookupTable(owner, app string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("LookupTable")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", lookupTablePath(owner, app, ""), params)
}

// ReadLookupTable reads a lookup table file metadata.
func (c *SplunkClient) ReadLookupTable(owner, app, name string) (map[string]interface{}, error) {
	return c.DoRequest("GET", lookupTablePath(owner, app, name), nil)
}

// UpdateLookupTable updates a lookup table file.
func (c *SplunkClient) UpdateLookupTable(owner, app, name string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("LookupTable")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", lookupTablePath(owner, app, name), params)
}

// DeleteLookupTable deletes a lookup table file.
func (c *SplunkClient) DeleteLookupTable(owner, app, name string) error {
	lock := c.GetResourceLock("LookupTable")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(lookupTablePath(owner, app, name))
}

// UpdateLookupTableContent updates the CSV content of a lookup table using the lookup_edit endpoint.
func (c *SplunkClient) UpdateLookupTableContent(owner, app, filename, contents string) (map[string]interface{}, error) {
	lock := c.GetResourceLock("LookupTable")
	lock.Lock()
	defer lock.Unlock()

	params := url.Values{}
	params.Set("namespace", app)
	params.Set("owner", owner)
	params.Set("lookup_file", filename)
	params.Set("contents", contents)

	return c.DoRequest("POST", "/services/data/lookup_edit/lookup_contents", params)
}
