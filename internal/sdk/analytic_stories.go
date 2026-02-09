package sdk

import (
	"fmt"
	"net/url"
)

// --- Analytic Stories CRUD ---
// Endpoint: /servicesNS/{owner}/{app}/configs/conf-analyticstories

func analyticStoryPath(owner, app, name string) string {
	if name != "" {
		return fmt.Sprintf("/servicesNS/%s/%s/configs/conf-analyticstories/%s", URLEncode(owner), URLEncode(app), URLEncode(name))
	}
	return fmt.Sprintf("/servicesNS/%s/%s/configs/conf-analyticstories", URLEncode(owner), URLEncode(app))
}

// CreateAnalyticStory creates a new analytic story.
func (c *SplunkClient) CreateAnalyticStory(owner, app string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("AnalyticStory")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", analyticStoryPath(owner, app, ""), params)
}

// ReadAnalyticStory reads an analytic story by name.
func (c *SplunkClient) ReadAnalyticStory(owner, app, name string) (map[string]interface{}, error) {
	return c.DoRequest("GET", analyticStoryPath(owner, app, name), nil)
}

// UpdateAnalyticStory updates an analytic story.
func (c *SplunkClient) UpdateAnalyticStory(owner, app, name string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("AnalyticStory")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", analyticStoryPath(owner, app, name), params)
}

// DeleteAnalyticStory deletes an analytic story.
func (c *SplunkClient) DeleteAnalyticStory(owner, app, name string) error {
	lock := c.GetResourceLock("AnalyticStory")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(analyticStoryPath(owner, app, name))
}
