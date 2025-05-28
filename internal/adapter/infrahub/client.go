package infrahub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/simli1333/vidra/internal/domain"
)

type infrahubClient struct{}

// NewClient returns a new InfrahubClient.
func NewClient() domain.InfrahubClient {
	return &infrahubClient{}
}

var relativeFormatRegex = regexp.MustCompile(`^now[-+]\d+[smhdw]$`)

// IsValidTargetDateFormat checks if the input string is a valid RFC3339 or relative time format.
func IsValidTargetDateFormat(input string) error {
	if _, err := time.Parse(time.RFC3339, input); err == nil {
		return nil
	}
	if relativeFormatRegex.MatchString(input) {
		return nil
	}
	return fmt.Errorf("targetDate must be RFC3339 or relative like 'now-2h', got: %s", input)
}

// BuildURL builds a URL using base API URL, path with placeholders, path parameters, and query parameters.
func BuildURL(baseAPIURL, pathTemplate string, pathParams map[string]string, queryParams map[string]string) (string, error) {
	// Replace placeholders in the path (e.g. :id)
	path := pathTemplate
	for key, val := range pathParams {
		placeholder := fmt.Sprintf(":%s", key)
		path = strings.ReplaceAll(path, placeholder, url.PathEscape(val))
	}

	// Prepare query parameters
	q := url.Values{}
	for k, v := range queryParams {
		if v == "" {
			continue // skip empty values
		}

		// Special validation for 'at' (targetDate)
		if k == "at" {
			if err := IsValidTargetDateFormat(v); err != nil {
				return "", fmt.Errorf("invalid 'at' query param format: %w", err)
			}
		}

		q.Set(k, v)
	}

	fullURL := fmt.Sprintf("%s%s", strings.TrimSuffix(baseAPIURL, "/"), path)

	if encodedQuery := q.Encode(); encodedQuery != "" {
		fullURL += "?" + encodedQuery
	}

	return fullURL, nil
}

// RunQuery sends a query to the Infrahub API
func (c *infrahubClient) RunQuery(queryName string, apiURL string, artifactName string, targetBranche string, targetDate string, token string) (*[]domain.Artifact, error) {
	// Construct the query URL
	url, err := BuildURL(
		apiURL,
		fmt.Sprintf("/api/query/%s", queryName),
		nil,
		map[string]string{
			"update_group": "false",
			"branch":       targetBranche,
			"at":           targetDate,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build query URL: %w", err)
	}

	payload := queryPayload{
		Variables: map[string]string{
			"artifactname": artifactName,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query payload: %w", err)
	}

	var resp *http.Response
	var lastErr error
	backoff := 200 * time.Millisecond

	for attempts := 0; attempts < 5; attempts++ {
		req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create query request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			if cerr := resp.Body.Close(); cerr != nil {
				fmt.Printf("warning: failed to close response body: %v\n", cerr)
			}
		}
		lastErr = err
		time.Sleep(backoff)
		backoff *= 2
	}

	if resp == nil {
		return nil, fmt.Errorf("query request failed after retries: %w", lastErr)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("warning: failed to close response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed with status %s: %s", resp.Status, body)
	}

	var queryResult artifactIDQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&queryResult); err != nil {
		return nil, fmt.Errorf("failed to decode query result: %w", err)
	}

	artifacts, err := CreateArtifactsFromAPIResponse(queryResult)
	if err != nil {
		return nil, fmt.Errorf("failed to create artifacts from API response: %w", err)
	}

	return &artifacts, nil
}

// Login authenticates with the Infrahub API and returns the authentication token
func (c *infrahubClient) Login(apiURL, username, password string) (string, error) {
	loginURL := fmt.Sprintf("%s/api/auth/login", apiURL)
	loginPayload := map[string]string{"username": username, "password": password}

	payloadBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login payload: %w", err)
	}

	var resp *http.Response
	var lastErr error
	backoff := time.Millisecond * 200

	for attempts := 0; attempts < 5; attempts++ {
		req, err := http.NewRequest("POST", loginURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return "", fmt.Errorf("failed to create login request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			if cerr := resp.Body.Close(); cerr != nil {
				fmt.Printf("warning: failed to close response body: %v\n", cerr)
			}
		}
		lastErr = err
		time.Sleep(backoff)
		backoff *= 2
	}

	if resp == nil {
		return "", fmt.Errorf("login request failed after retries: %w", lastErr)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %s: %s", resp.Status, body)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	return loginResp.Token, nil
}

// DownloadArtifact downloads the artifact from the given URL and saves it to a temporary file
func (c *infrahubClient) DownloadArtifact(apiURL string, artifactID string, targetBranche string, targetDate string, token string) (io.Reader, error) {
	url, err := BuildURL(
		apiURL,
		"/api/artifact/:artifactID",
		map[string]string{
			"artifactID": artifactID,
		},
		map[string]string{
			"branch": targetBranche,
			"at":     targetDate,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL for artifact: %v", err)
	}

	var resp *http.Response
	var lastErr error
	backoff := 200 * time.Millisecond

	for attempts := 0; attempts < 5; attempts++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err = http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			// Success
			return resp.Body, nil
		}
		if resp != nil {
			if cerr := resp.Body.Close(); cerr != nil {
				fmt.Printf("warning: failed to close response body: %v\n", cerr)
			}
		}
		lastErr = err
		time.Sleep(backoff)
		backoff *= 2
	}

	if resp != nil {
		body, _ := io.ReadAll(resp.Body)
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("warning: failed to close response body: %v\n", cerr)
		}
		return nil, fmt.Errorf("failed to download artifact after retries, last status code: %d, response: %s", resp.StatusCode, body)
	}
	return nil, fmt.Errorf("failed to download artifact after retries: %v", lastErr)
}
