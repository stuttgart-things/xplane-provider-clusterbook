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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("plain HTTP", func(t *testing.T) {
		c, err := NewClient("http://localhost:8080", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.baseURL != "http://localhost:8080" {
			t.Errorf("baseURL = %q, want %q", c.baseURL, "http://localhost:8080")
		}
	})

	t.Run("trims trailing slash", func(t *testing.T) {
		c, err := NewClient("http://localhost:8080/", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.baseURL != "http://localhost:8080" {
			t.Errorf("baseURL = %q, want %q", c.baseURL, "http://localhost:8080")
		}
	})

	t.Run("insecure skip verify", func(t *testing.T) {
		c, err := NewClient("https://localhost:8443", &TLSOptions{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("client is nil")
		}
	})

	t.Run("invalid custom CA", func(t *testing.T) {
		_, err := NewClient("https://localhost:8443", &TLSOptions{CustomCA: "not-a-cert"})
		if err == nil {
			t.Fatal("expected error for invalid CA")
		}
	})
}

func TestReserveIPs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/api/v1/networks/10.31.103/reserve" {
				t.Errorf("path = %s, want /api/v1/networks/10.31.103/reserve", r.URL.Path)
			}

			var req ReserveRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("cannot decode request body: %v", err)
			}
			if req.Cluster != "mycluster" || req.Count != 1 {
				t.Errorf("request = %+v, want cluster=mycluster count=1", req)
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(ReserveResponse{IPs: []string{"10.31.103.10"}, Status: "ASSIGNED"}) //nolint:errcheck
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		resp, err := c.ReserveIPs(context.Background(), "10.31.103", ReserveRequest{Cluster: "mycluster", Count: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.IPs) != 1 || resp.IPs[0] != "10.31.103.10" {
			t.Errorf("IPs = %v, want [10.31.103.10]", resp.IPs)
		}
		if resp.Status != "ASSIGNED" {
			t.Errorf("Status = %q, want ASSIGNED", resp.Status)
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error")) //nolint:errcheck
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		_, err := c.ReserveIPs(context.Background(), "10.31.103", ReserveRequest{Cluster: "mycluster", Count: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetIPs(t *testing.T) {
	t.Run("returns IPs without FQDN", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/networks/10.31.103/ips" {
				t.Errorf("path = %s", r.URL.Path)
			}
			json.NewEncoder(w).Encode([]IPInfo{ //nolint:errcheck
				{IP: "10.31.103.10", Status: "ASSIGNED", Cluster: "mycluster"},
			})
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		ips, err := c.GetIPs(context.Background(), "10.31.103")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ips) != 1 || ips[0].IP != "10.31.103.10" {
			t.Errorf("IPs = %+v", ips)
		}
		if ips[0].FQDN != "" {
			t.Errorf("FQDN should be empty for non-DNS IP, got %q", ips[0].FQDN)
		}
	})

	t.Run("returns IPs with FQDN for DNS entries", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			json.NewEncoder(w).Encode([]IPInfo{ //nolint:errcheck
				{IP: "10.31.103.10", Status: "ASSIGNED:DNS", Cluster: "mycluster", FQDN: "*.mycluster.example.com"},
				{IP: "10.31.103.11", Status: "ASSIGNED", Cluster: "other"},
			})
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		ips, err := c.GetIPs(context.Background(), "10.31.103")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ips) != 2 {
			t.Fatalf("got %d IPs, want 2", len(ips))
		}
		if ips[0].FQDN != "*.mycluster.example.com" {
			t.Errorf("FQDN = %q, want *.mycluster.example.com", ips[0].FQDN)
		}
		if ips[1].FQDN != "" {
			t.Errorf("non-DNS IP should have no FQDN, got %q", ips[1].FQDN)
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		_, err := c.GetIPs(context.Background(), "10.31.103")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetClusterInfo(t *testing.T) {
	t.Run("returns FQDN and zone with DNS", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/clusters/mycluster" {
				t.Errorf("path = %s, want /api/v1/clusters/mycluster", r.URL.Path)
			}
			json.NewEncoder(w).Encode(ClusterInfo{ //nolint:errcheck
				Cluster: "mycluster",
				FQDN:    "*.mycluster.example.com",
				Zone:    "example.com",
			})
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		info, err := c.GetClusterInfo(context.Background(), "mycluster")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Cluster != "mycluster" {
			t.Errorf("Cluster = %q, want mycluster", info.Cluster)
		}
		if info.FQDN != "*.mycluster.example.com" {
			t.Errorf("FQDN = %q, want *.mycluster.example.com", info.FQDN)
		}
		if info.Zone != "example.com" {
			t.Errorf("Zone = %q, want example.com", info.Zone)
		}
	})

	t.Run("returns cluster without DNS fields", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			json.NewEncoder(w).Encode(ClusterInfo{Cluster: "nodnsc"}) //nolint:errcheck
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		info, err := c.GetClusterInfo(context.Background(), "nodnsc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.FQDN != "" {
			t.Errorf("FQDN should be empty, got %q", info.FQDN)
		}
		if info.Zone != "" {
			t.Errorf("Zone should be empty, got %q", info.Zone)
		}
	})

	t.Run("not found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("cluster not found")) //nolint:errcheck
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		_, err := c.GetClusterInfo(context.Background(), "missing")
		if err == nil {
			t.Fatal("expected error for missing cluster")
		}
	})
}

func TestReleaseIPs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/api/v1/networks/10.31.103/release" {
				t.Errorf("path = %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		err := c.ReleaseIPs(context.Background(), "10.31.103", ReleaseRequest{IP: "10.31.103.10"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		err := c.ReleaseIPs(context.Background(), "10.31.103", ReleaseRequest{IP: "10.31.103.10"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUpdateIP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("method = %s, want PUT", r.Method)
			}
			if r.URL.Path != "/api/v1/networks/10.31.103/ips/10.31.103.10" {
				t.Errorf("path = %s", r.URL.Path)
			}

			var req ReserveRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("cannot decode: %v", err)
			}
			if !req.CreateDNS {
				t.Error("createDNS should be true")
			}
			if req.Status != "ASSIGNED" {
				t.Errorf("status = %q, want ASSIGNED", req.Status)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		err := c.UpdateIP(context.Background(), "10.31.103", "10.31.103.10", ReserveRequest{
			Cluster:   "mycluster",
			CreateDNS: true,
			Status:    "ASSIGNED",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c, _ := NewClient(srv.URL, nil)
		err := c.UpdateIP(context.Background(), "10.31.103", "10.31.103.10", ReserveRequest{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
