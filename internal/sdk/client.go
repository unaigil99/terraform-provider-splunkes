package sdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// SplunkClient is the HTTP client for Splunk Enterprise and ES REST APIs.
type SplunkClient struct {
	BaseURL            string
	Username           string
	Password           string
	AuthToken          string
	SessionKey         string
	InsecureSkipVerify bool
	HTTPClient         *http.Client
	Timeout            int

	mu             sync.Mutex
	resourceLocks  map[string]*sync.Mutex
	rateLimiter    <-chan time.Time
}

// NewClient creates a new SplunkClient and authenticates.
func NewClient(baseURL, username, password, authToken string, insecureSkipVerify bool, timeout int) (*SplunkClient, error) {
	if timeout <= 0 {
		timeout = 60
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	c := &SplunkClient{
		BaseURL:            strings.TrimRight(baseURL, "/"),
		Username:           username,
		Password:           password,
		AuthToken:          authToken,
		InsecureSkipVerify: insecureSkipVerify,
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeout) * time.Second,
		},
		Timeout:       timeout,
		resourceLocks: make(map[string]*sync.Mutex),
		rateLimiter:   time.Tick(200 * time.Millisecond),
	}

	// If no bearer token, authenticate with username/password to get session key
	if c.AuthToken == "" && c.Username != "" && c.Password != "" {
		err := c.login()
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	return c, nil
}

// login authenticates with username/password and stores the session key.
func (c *SplunkClient) login() error {
	data := url.Values{}
	data.Set("username", c.Username)
	data.Set("password", c.Password)
	data.Set("output_mode", "json")

	req, err := http.NewRequest("POST", c.BaseURL+"/services/auth/login", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if sessionKey, ok := result["sessionKey"].(string); ok {
		c.SessionKey = sessionKey
	} else {
		return fmt.Errorf("sessionKey not found in login response")
	}

	return nil
}

// GetResourceLock returns a mutex for the given resource type.
func (c *SplunkClient) GetResourceLock(resourceType string) *sync.Mutex {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.resourceLocks[resourceType]; !ok {
		c.resourceLocks[resourceType] = &sync.Mutex{}
	}
	return c.resourceLocks[resourceType]
}

// setAuthHeader sets the appropriate authorization header.
func (c *SplunkClient) setAuthHeader(req *http.Request) {
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	} else if c.SessionKey != "" {
		req.Header.Set("Authorization", "Splunk "+c.SessionKey)
	} else if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}
}

// DoRequest executes an HTTP request against the Splunk REST API.
// It handles authentication, rate limiting, and retries for 429/5xx.
func (c *SplunkClient) DoRequest(method, path string, body url.Values) (map[string]interface{}, error) {
	<-c.rateLimiter

	fullURL := c.BaseURL + path
	if !strings.Contains(path, "output_mode=") {
		sep := "?"
		if strings.Contains(path, "?") {
			sep = "&"
		}
		fullURL += sep + "output_mode=json"
	}

	var encodedBody string
	if body != nil && (method == "POST" || method == "PUT") {
		encodedBody = body.Encode()
	}

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		var reqBody io.Reader
		if encodedBody != "" {
			reqBody = strings.NewReader(encodedBody)
		}
		req, err := http.NewRequest(method, fullURL, reqBody)
		if err != nil {
			return nil, err
		}
		if body != nil && (method == "POST" || method == "PUT") {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		c.setAuthHeader(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode == 429 {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}

		if resp.StatusCode >= 500 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error %d on %s %s: %s", resp.StatusCode, method, path, string(respBody))
		}

		var result map[string]interface{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, &result); err != nil {
				return nil, fmt.Errorf("parsing response JSON: %w (body: %s)", err, string(respBody))
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("request failed after 5 retries: %w", lastErr)
}

// DoJSONRequest executes an HTTP request with a JSON body (used for ES v2 API).
func (c *SplunkClient) DoJSONRequest(method, path string, jsonBody interface{}) (map[string]interface{}, error) {
	<-c.rateLimiter

	fullURL := c.BaseURL + path

	var jsonData []byte
	if jsonBody != nil {
		var err error
		jsonData, err = json.Marshal(jsonBody)
		if err != nil {
			return nil, fmt.Errorf("marshaling JSON body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		var reqBody io.Reader
		if jsonData != nil {
			reqBody = bytes.NewReader(jsonData)
		}
		req, err := http.NewRequest(method, fullURL, reqBody)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		c.setAuthHeader(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode == 429 {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}

		if resp.StatusCode >= 500 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error %d on %s %s: %s", resp.StatusCode, method, path, string(respBody))
		}

		var result map[string]interface{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, &result); err != nil {
				return nil, fmt.Errorf("parsing response JSON: %w (body: %s)", err, string(respBody))
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("request failed after 5 retries: %w", lastErr)
}

// DoDelete executes a DELETE request.
func (c *SplunkClient) DoDelete(path string) error {
	<-c.rateLimiter

	fullURL := c.BaseURL + path + "?output_mode=json"

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s returned %d: %s", path, resp.StatusCode, string(body))
	}

	return nil
}

// --- Helper functions for parsing API responses ---

// GetEntryContent extracts the first entry's content from a standard Splunk REST response.
func GetEntryContent(response map[string]interface{}) (map[string]interface{}, error) {
	entries, ok := response["entry"].([]interface{})
	if !ok || len(entries) == 0 {
		return nil, fmt.Errorf("no entries found in response")
	}
	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid entry format")
	}
	content, ok := entry["content"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no content in entry")
	}
	return content, nil
}

// GetEntryACL extracts the first entry's ACL from a standard Splunk REST response.
func GetEntryACL(response map[string]interface{}) (map[string]interface{}, error) {
	entries, ok := response["entry"].([]interface{})
	if !ok || len(entries) == 0 {
		return nil, fmt.Errorf("no entries found in response")
	}
	entry, ok := entries[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid entry format")
	}
	acl, ok := entry["acl"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no acl in entry")
	}
	return acl, nil
}

// ParseString safely extracts a string value from a map.
func ParseString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok && v != nil {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return fmt.Sprintf("%g", val)
		case bool:
			if val {
				return "1"
			}
			return "0"
		default:
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// ParseBool safely extracts a boolean value from a map.
func ParseBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key]; ok && v != nil {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "1" || strings.EqualFold(val, "true")
		case float64:
			return val != 0
		}
	}
	return false
}

// ParseInt safely extracts an integer value from a map.
func ParseInt(data map[string]interface{}, key string) int64 {
	if v, ok := data[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return int64(val)
		case string:
			var i int64
			fmt.Sscanf(val, "%d", &i)
			return i
		}
	}
	return 0
}

// URLEncode encodes a string for use in a URL path segment.
func URLEncode(s string) string {
	return url.PathEscape(s)
}
