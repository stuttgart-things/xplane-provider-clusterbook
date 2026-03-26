/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an HTTP client for the clusterbook REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new clusterbook API client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IPInfo represents an IP address entry returned by the clusterbook API.
type IPInfo struct {
	IP          string `json:"ip"`
	Network     string `json:"network"`
	ClusterName string `json:"clusterName"`
	Status      string `json:"status"`
}

// ReserveRequest is the request body for reserving IPs.
type ReserveRequest struct {
	ClusterName string `json:"clusterName"`
	Count       int    `json:"count,omitempty"`
	IP          string `json:"ip,omitempty"`
	CreateDNS   bool   `json:"createDNS,omitempty"`
}

// ReserveResponse is the response from the reserve/assign endpoint.
type ReserveResponse struct {
	IPs    []string `json:"ips"`
	Status string   `json:"status"`
}

// ReleaseRequest is the request body for releasing IPs.
type ReleaseRequest struct {
	ClusterName string `json:"clusterName"`
}

// ReserveIPs reserves IPs from the given network pool.
func (c *Client) ReserveIPs(ctx context.Context, networkKey string, req ReserveRequest) (*ReserveResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal reserve request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/networks/%s/reserve", c.baseURL, networkKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create reserve request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("reserve request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reserve request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ReserveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot decode reserve response: %w", err)
	}
	return &result, nil
}

// GetIPs returns the IPs assigned to a cluster in the given network.
func (c *Client) GetIPs(ctx context.Context, networkKey string) ([]IPInfo, error) {
	url := fmt.Sprintf("%s/api/v1/networks/%s/ips", c.baseURL, networkKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create get IPs request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("get IPs request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get IPs returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ips []IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&ips); err != nil {
		return nil, fmt.Errorf("cannot decode IPs response: %w", err)
	}
	return ips, nil
}

// ReleaseIPs releases IPs assigned to a cluster in the given network.
func (c *Client) ReleaseIPs(ctx context.Context, networkKey string, req ReleaseRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("cannot marshal release request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/networks/%s/release", c.baseURL, networkKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create release request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("release request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("release request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateIP updates an existing IP assignment.
func (c *Client) UpdateIP(ctx context.Context, networkKey, ip string, req ReserveRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("cannot marshal update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/networks/%s/ips/%s", c.baseURL, networkKey, ip)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create update request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("update request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
