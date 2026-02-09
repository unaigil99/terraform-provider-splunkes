package sdk

import (
	"fmt"
	"net/url"
)

// --- KV Store Collections CRUD ---
// Endpoint: /servicesNS/{owner}/{app}/storage/collections/config

func kvstoreConfigPath(owner, app, collection string) string {
	if collection != "" {
		return fmt.Sprintf("/servicesNS/%s/%s/storage/collections/config/%s", URLEncode(owner), URLEncode(app), URLEncode(collection))
	}
	return fmt.Sprintf("/servicesNS/%s/%s/storage/collections/config", URLEncode(owner), URLEncode(app))
}

// CreateKVStoreCollection creates a new KV store collection.
func (c *SplunkClient) CreateKVStoreCollection(owner, app string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("KVStore")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", kvstoreConfigPath(owner, app, ""), params)
}

// ReadKVStoreCollection reads a KV store collection configuration.
func (c *SplunkClient) ReadKVStoreCollection(owner, app, collection string) (map[string]interface{}, error) {
	return c.DoRequest("GET", kvstoreConfigPath(owner, app, collection), nil)
}

// UpdateKVStoreCollection updates a KV store collection configuration.
func (c *SplunkClient) UpdateKVStoreCollection(owner, app, collection string, params url.Values) (map[string]interface{}, error) {
	lock := c.GetResourceLock("KVStore")
	lock.Lock()
	defer lock.Unlock()

	return c.DoRequest("POST", kvstoreConfigPath(owner, app, collection), params)
}

// DeleteKVStoreCollection deletes a KV store collection.
func (c *SplunkClient) DeleteKVStoreCollection(owner, app, collection string) error {
	lock := c.GetResourceLock("KVStore")
	lock.Lock()
	defer lock.Unlock()

	return c.DoDelete(kvstoreConfigPath(owner, app, collection))
}
