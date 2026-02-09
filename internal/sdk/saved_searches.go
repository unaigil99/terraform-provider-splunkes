package sdk

import (
	"fmt"
	"net/url"
)

// --- Saved Searches CRUD (standard Splunk REST API) ---
// Endpoint: /servicesNS/{owner}/{app}/saved/searches

func savedSearchPath(owner, app, name string) string {
	if name != "" {
		return fmt.Sprintf("/servicesNS/%s/%s/saved/searches/%s", URLEncode(owner), URLEncode(app), URLEncode(name))
	}
	return fmt.Sprintf("/servicesNS/%s/%s/saved/searches", URLEncode(owner), URLEncode(app))
}

// CreateSavedSearch creates a new saved search.
func (c *SplunkClient) CreateSavedSearch(owner, app string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("SavedSearch")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", savedSearchPath(owner, app, ""), params)
}

// ReadSavedSearch reads a saved search by name.
func (c *SplunkClient) ReadSavedSearch(owner, app, name string) (map[string]interface{}, error) {
	return c.DoRequest("GET", savedSearchPath(owner, app, name), nil)
}

// UpdateSavedSearch updates a saved search by name.
func (c *SplunkClient) UpdateSavedSearch(owner, app, name string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("SavedSearch")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", savedSearchPath(owner, app, name), params)
}

// DeleteSavedSearch deletes a saved search by name.
func (c *SplunkClient) DeleteSavedSearch(owner, app, name string) error {
	lock := c.GetResourceLock("SavedSearch")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(savedSearchPath(owner, app, name))
}

// ListSavedSearches lists saved searches with optional search filter.
func (c *SplunkClient) ListSavedSearches(owner, app, filter string) (map[string]interface{}, error) {
	path := savedSearchPath(owner, app, "")
	if filter != "" {
		path += "?count=-1&search=" + url.QueryEscape(filter)
	} else {
		path += "?count=-1"
	}
	return c.DoRequest("GET", path, nil)
}
