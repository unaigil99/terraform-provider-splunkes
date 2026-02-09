package sdk

import (
	"fmt"
	"net/url"
)

// --- Macros CRUD ---
// Endpoint: /servicesNS/{owner}/{app}/configs/conf-macros

func macroPath(owner, app, name string) string {
	if name != "" {
		return fmt.Sprintf("/servicesNS/%s/%s/configs/conf-macros/%s", URLEncode(owner), URLEncode(app), URLEncode(name))
	}
	return fmt.Sprintf("/servicesNS/%s/%s/configs/conf-macros", URLEncode(owner), URLEncode(app))
}

// CreateMacro creates a new search macro.
func (c *SplunkClient) CreateMacro(owner, app string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("Macro")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", macroPath(owner, app, ""), params)
}

// ReadMacro reads a search macro by name.
func (c *SplunkClient) ReadMacro(owner, app, name string) (map[string]interface{}, error) {
	return c.DoRequest("GET", macroPath(owner, app, name), nil)
}

// UpdateMacro updates a search macro by name.
func (c *SplunkClient) UpdateMacro(owner, app, name string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("Macro")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", macroPath(owner, app, name), params)
}

// DeleteMacro deletes a search macro by name.
func (c *SplunkClient) DeleteMacro(owner, app, name string) error {
	lock := c.GetResourceLock("Macro")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(macroPath(owner, app, name))
}

// ListMacros lists all macros in a given app context.
func (c *SplunkClient) ListMacros(owner, app string) (map[string]interface{}, error) {
	return c.DoRequest("GET", macroPath(owner, app, "")+"?count=-1", nil)
}
