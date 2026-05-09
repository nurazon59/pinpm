package main

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	pkg, err := LoadPackageJSON(path)
	require.NoError(t, err)

	assert.Equal(t, "test-pkg", pkg.Name)
	assert.Equal(t, "^4.0.0", pkg.Dependencies["express"])
	assert.Equal(t, "^29.0.0", pkg.DevDependencies["jest"])
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

	require.NoError(t, SavePackageJSON(path, pkg))

	loaded, err := LoadPackageJSON(path)
	require.NoError(t, err)

	assert.Equal(t, "test-pkg", loaded.Name)
	assert.Equal(t, "4.21.0", loaded.Dependencies["express"])
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
	assert.Len(t, deps, 2)

	depMap := make(map[string]Dependency)
	for _, d := range deps {
		depMap[d.Name] = d
	}

	assert.Equal(t, "dependencies", depMap["express"].Section)
	assert.Equal(t, "devDependencies", depMap["jest"].Section)
}

func TestUpdateDependency(t *testing.T) {
	pkg := &PackageJSON{}

	require.NoError(t, pkg.setDependency("dependencies", "express", "4.21.0"))
	assert.Equal(t, "4.21.0", pkg.Dependencies["express"])

	assert.Error(t, pkg.setDependency("invalid", "pkg", "1.0.0"))
}

func TestUnknownFieldsPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	original := `{
		"name": "test-pkg",
		"version": "1.0.0",
		"scripts": {
			"start": "node index.js",
			"test": "jest"
		},
		"license": "MIT",
		"description": "Test package",
		"dependencies": {"express": "^4.0.0"}
	}`

	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	pkg, err := LoadPackageJSON(path)
	require.NoError(t, err)

	require.NoError(t, SavePackageJSON(path, pkg))

	loaded, err := LoadPackageJSON(path)
	require.NoError(t, err)

	assert.NotEmpty(t, string(loaded.Extra["scripts"]))
	assert.NotEmpty(t, string(loaded.Extra["license"]))
	assert.NotEmpty(t, string(loaded.Extra["description"]))
	assert.Equal(t, "^4.0.0", loaded.Dependencies["express"])
}

func TestScriptsAmpersandPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	original := `{
		"name": "test-pkg",
		"version": "1.0.0",
		"scripts": {
			"build": "tsc && node dist/index.js",
			"test": "echo 'hello && world'"
		},
		"dependencies": {"express": "^4.0.0"}
	}`

	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	pkg, err := LoadPackageJSON(path)
	require.NoError(t, err)

	require.NoError(t, SavePackageJSON(path, pkg))

	loaded, err := LoadPackageJSON(path)
	require.NoError(t, err)

	var originalScripts, loadedScripts map[string]string
	require.NoError(t, json.Unmarshal(pkg.Extra["scripts"], &originalScripts))
	require.NoError(t, json.Unmarshal(loaded.Extra["scripts"], &loadedScripts))

	assert.True(t, maps.Equal(originalScripts, loadedScripts))
}
