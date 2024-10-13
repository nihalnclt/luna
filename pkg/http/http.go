package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nihalnclt/luna/models"
)

const REGISTRY_URL = "https://registry.npmjs.org"

var client = &http.Client{}

// Registry fetches the detail of a package including the versions and dependencies details.
func Registry(route string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s%s", REGISTRY_URL, route), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Recommended header to shorten the response size.
	req.Header.Set("Accept", "application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8, */*")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return buf.String(), nil
}

// VersionData retrieves information for a specific version of a package.
// This method is preferred due to its reduced response size.
func VersionData(packageName string, version string) (*models.VersionData, error) {
	// example (/latest, /1.0.0)
	route := fmt.Sprintf("/%s/%s", packageName, version)
	response, err := Registry(route)
	if err != nil {
		return nil, err
	}

	var versionData models.VersionData
	if err := json.Unmarshal([]byte(response), &versionData); err != nil {
		return nil, fmt.Errorf("failed unmarshal version data: %w", err)
	}

	return &versionData, nil
}

// PackageData retrieves the package with all versions
func PackageData(packageName string) (*models.PackageData, error) {
	route := fmt.Sprintf("/%s", packageName)
	response, err := Registry(route)
	if err != nil {
		return nil, err
	}

	var packageData models.PackageData
	if err := json.Unmarshal([]byte(response), &packageData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package data: %w", err)
	}

	return &packageData, nil
}

// GetBytes download file from any specified url
func GetBytes(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return buf.Bytes(), nil
}
