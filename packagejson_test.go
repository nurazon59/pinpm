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
	tests := map[string]struct {
		content     string
		wantName    string
		wantExpress string
		wantJest    string
	}{
		"basic": {
			content: `{
				"name": "test-pkg",
				"version": "1.0.0",
				"dependencies": {"express": "^4.0.0"},
				"devDependencies": {"jest": "^29.0.0"}
			}`,
			wantName:    "test-pkg",
			wantExpress: "^4.0.0",
			wantJest:    "^29.0.0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "package.json")
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o644))

			pkg, err := LoadPackageJSON(path)
			require.NoError(t, err)

			assert.Equal(t, tt.wantName, pkg.Name)
			assert.Equal(t, tt.wantExpress, pkg.Dependencies["express"])
			assert.Equal(t, tt.wantJest, pkg.DevDependencies["jest"])
		})
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
	tests := map[string]struct {
		section string
		name    string
		version string
		wantErr bool
		want    string
	}{
		"dependencies": {
			section: "dependencies",
			name:    "express",
			version: "4.21.0",
			want:    "4.21.0",
		},
		"invalid section": {
			section: "invalid",
			name:    "pkg",
			version: "1.0.0",
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pkg := &PackageJSON{}
			err := pkg.setDependency(tt.section, tt.name, tt.version)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, pkg.Dependencies["express"])
		})
	}
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

func TestKeysAreSorted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	pkg := &PackageJSON{
		Name:    "test-pkg",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"zebra":  "^1.0.0",
			"alpha":  "^2.0.0",
			"middle": "^3.0.0",
		},
		DevDependencies: map[string]string{
			"jest":   "^29.0.0",
			"eslint": "^8.0.0",
		},
	}

	require.NoError(t, SavePackageJSON(path, pkg))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)

	depIdx := indexOf(content, "dependencies")
	devIdx := indexOf(content, "devDependencies")
	nameIdx := indexOf(content, "name")
	versionIdx := indexOf(content, "version")

	assert.Less(t, nameIdx, versionIdx, "name should come before version")
	assert.Less(t, depIdx, devIdx, "dependencies should come before devDependencies")

	depContent := content[depIdx:]
	alphaIdx := indexOf(depContent, "alpha")
	middleIdx := indexOf(depContent, "middle")
	zebraIdx := indexOf(depContent, "zebra")

	assert.Less(t, alphaIdx, middleIdx, "alpha should come before middle")
	assert.Less(t, middleIdx, zebraIdx, "middle should come before zebra")
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
