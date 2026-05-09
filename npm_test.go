package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPackage(t *testing.T) {
	tests := map[string]struct {
		path         string
		responseBody string
		responseCode int
		packageName  string
		wantName     string
		wantLatest   string
		wantErr      bool
	}{
		"found": {
			path:         "/express",
			responseCode: http.StatusOK,
			responseBody: `{
				"name": "express",
				"dist-tags": {"latest": "4.21.0"},
				"versions": {"4.21.0": {"version": "4.21.0"}}
			}`,
			packageName: "express",
			wantName:    "express",
			wantLatest:  "4.21.0",
		},
		"not found": {
			path:         "/nonexistent",
			responseCode: http.StatusNotFound,
			responseBody: "",
			packageName:  "nonexistent",
			wantErr:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			pkg, err := client.GetPackage(tt.packageName)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, pkg.Name)
			assert.Equal(t, tt.wantLatest, pkg.DistTags["latest"])
		})
	}
}

func TestResolveVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"name": "test-pkg",
			"dist-tags": {"latest": "3.0.0"},
			"versions": {
				"1.0.0": {},
				"1.1.0": {},
				"1.2.0": {},
				"2.0.0": {},
				"2.1.0": {},
				"3.0.0": {}
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	tests := map[string]struct {
		rangeSpec string
		want      string
	}{
		"caret major": {"^1.0.0", "1.2.0"},
		"tilde minor": {"~1.1.0", "1.1.0"},
		"exact":       {"2.0.0", "2.0.0"},
		"gte":         {">=2.0.0", "3.0.0"},
		"wildcard":    {"*", "3.0.0"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := client.ResolveVersion("test-pkg", tt.rangeSpec)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPackageNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetPackage("nonexistent")
	assert.Error(t, err)
}
