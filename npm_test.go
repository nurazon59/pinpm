package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/express" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"name": "express",
			"dist-tags": {"latest": "4.21.0"},
			"versions": {"4.21.0": {"version": "4.21.0"}}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	pkg, err := client.GetPackage("express")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Name != "express" {
		t.Errorf("expected name 'express', got %q", pkg.Name)
	}
	if pkg.DistTags["latest"] != "4.21.0" {
		t.Errorf("expected latest '4.21.0', got %q", pkg.DistTags["latest"])
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveVersion(%q) = %q, want %q", tt.rangeSpec, got, tt.want)
			}
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
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
