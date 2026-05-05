package pinpm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	err := os.WriteFile(path, []byte(`{
		"name": "test-pkg",
		"version": "1.0.0",
		"dependencies": {"express": "^4.0.0"},
		"devDependencies": {"jest": "^29.0.0"}
	}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	pkg, err := LoadPackageJSON(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Name != "test-pkg" {
		t.Errorf("expected name 'test-pkg', got %q", pkg.Name)
	}
	if pkg.Dependencies["express"] != "^4.0.0" {
		t.Errorf("expected express '^4.0.0', got %q", pkg.Dependencies["express"])
	}
	if pkg.DevDependencies["jest"] != "^29.0.0" {
		t.Errorf("expected jest '^29.0.0', got %q", pkg.DevDependencies["jest"])
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	pkg := &PackageJSON{
		Name:    "test-pkg",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"express": "4.21.0",
		},
	}

	if err := SavePackageJSON(path, pkg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := LoadPackageJSON(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Name != "test-pkg" {
		t.Errorf("expected name 'test-pkg', got %q", loaded.Name)
	}
	if loaded.Dependencies["express"] != "4.21.0" {
		t.Errorf("expected express '4.21.0', got %q", loaded.Dependencies["express"])
	}
}

func TestAllDependencies(t *testing.T) {
	pkg := &PackageJSON{
		Dependencies: map[string]string{
			"express": "^4.0.0",
		},
		DevDependencies: map[string]string{
			"jest": "^29.0.0",
		},
	}

	deps := pkg.AllDependencies()
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}

	depMap := make(map[string]Dependency)
	for _, d := range deps {
		depMap[d.Name] = d
	}

	if depMap["express"].Section != "dependencies" {
		t.Errorf("expected express in 'dependencies', got %q", depMap["express"].Section)
	}
	if depMap["jest"].Section != "devDependencies" {
		t.Errorf("expected jest in 'devDependencies', got %q", depMap["jest"].Section)
	}
}

func TestUpdateDependency(t *testing.T) {
	pkg := &PackageJSON{}

	if err := pkg.setDependency("dependencies", "express", "4.21.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkg.Dependencies["express"] != "4.21.0" {
		t.Errorf("expected express '4.21.0', got %q", pkg.Dependencies["express"])
	}

	if err := pkg.setDependency("invalid", "pkg", "1.0.0"); err == nil {
		t.Error("expected error for invalid section")
	}
}
