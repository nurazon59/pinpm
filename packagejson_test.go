package main

import (
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
			assert.Equal(t, tt.wantExpress, pkg.Dependencies.Values["express"])
			assert.Equal(t, tt.wantJest, pkg.DevDependencies.Values["jest"])
		})
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	deps := NewOrderedMap()
	deps.Set("express", "4.21.0")

	pkg := &PackageJSON{
		Name:         "test-pkg",
		Version:      "1.0.0",
		Dependencies: deps,
	}

	require.NoError(t, SavePackageJSON(path, pkg))

	loaded, err := LoadPackageJSON(path)
	require.NoError(t, err)

	assert.Equal(t, "test-pkg", loaded.Name)
	assert.Equal(t, "4.21.0", loaded.Dependencies.Values["express"])
}

func TestAllDependencies(t *testing.T) {
	deps := NewOrderedMap()
	deps.Set("express", "^4.0.0")

	devDeps := NewOrderedMap()
	devDeps.Set("jest", "^29.0.0")

	pkg := &PackageJSON{
		Dependencies:    deps,
		DevDependencies: devDeps,
	}

	depsList := pkg.AllDependencies()
	assert.Len(t, depsList, 2)

	depMap := make(map[string]Dependency)
	for _, d := range depsList {
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
			assert.Equal(t, tt.want, pkg.Dependencies.Values["express"])
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
	assert.Equal(t, "^4.0.0", loaded.Dependencies.Values["express"])
}

func TestScriptsAmpersandPreserved(t *testing.T) {
	tests := map[string]struct {
		scriptValue    string
		wantNotEscaped string
	}{
		"ampersand": {
			scriptValue:    "tsc && node dist/index.js",
			wantNotEscaped: "&&",
		},
		"less_than": {
			scriptValue:    "echo 'a < b'",
			wantNotEscaped: "<",
		},
		"greater_than": {
			scriptValue:    "echo 'a > b'",
			wantNotEscaped: ">",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "package.json")
			original := `{
				"name": "test-pkg",
				"version": "1.0.0",
				"scripts": {
					"build": "` + tt.scriptValue + `"
				},
				"dependencies": {"express": "^4.0.0"}
			}`

			require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

			pkg, err := LoadPackageJSON(path)
			require.NoError(t, err)

			require.NoError(t, SavePackageJSON(path, pkg))

			data, err := os.ReadFile(path)
			require.NoError(t, err)

			content := string(data)
			assert.Contains(t, content, tt.wantNotEscaped, tt.wantNotEscaped+" should not be escaped")
			assert.NotContains(t, content, `\u0026`, "should not contain escaped unicode")
		})
	}
}

func TestKeysPreserveOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	original := `{
		"name": "test-pkg",
		"version": "1.0.0",
		"dependencies": {
			"zebra": "^1.0.0",
			"alpha": "^2.0.0",
			"middle": "^3.0.0"
		},
		"devDependencies": {
			"jest": "^29.0.0",
			"eslint": "^8.0.0"
		}
	}`

	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	pkg, err := LoadPackageJSON(path)
	require.NoError(t, err)

	require.NoError(t, SavePackageJSON(path, pkg))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)

	depContent := content[indexOf(content, "dependencies"):]
	zebraIdx := indexOf(depContent, "zebra")
	alphaIdx := indexOf(depContent, "alpha")
	middleIdx := indexOf(depContent, "middle")

	assert.Less(t, zebraIdx, alphaIdx, "zebra should come before alpha (original order)")
	assert.Less(t, alphaIdx, middleIdx, "alpha should come before middle (original order)")

	devContent := content[indexOf(content, "devDependencies"):]
	jestIdx := indexOf(devContent, "jest")
	eslintIdx := indexOf(devContent, "eslint")

	assert.Less(t, jestIdx, eslintIdx, "jest should come before eslint (original order)")
}

func TestNewKeyAddedToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	original := `{
		"name": "test-pkg",
		"version": "1.0.0",
		"dependencies": {
			"alpha": "^1.0.0",
			"zebra": "^2.0.0"
		}
	}`

	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	pkg, err := LoadPackageJSON(path)
	require.NoError(t, err)

	pkg.setDependency("dependencies", "middle", "^3.0.0")

	require.NoError(t, SavePackageJSON(path, pkg))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)

	depContent := content[indexOf(content, "dependencies"):]
	alphaIdx := indexOf(depContent, "alpha")
	zebraIdx := indexOf(depContent, "zebra")
	middleIdx := indexOf(depContent, "middle")

	assert.Less(t, alphaIdx, zebraIdx, "alpha should come before zebra")
	assert.Less(t, zebraIdx, middleIdx, "new key 'middle' should be added at the end")
}

func TestExistingKeyUpdatedInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	original := `{
		"name": "test-pkg",
		"version": "1.0.0",
		"dependencies": {
			"alpha": "^1.0.0",
			"zebra": "^2.0.0"
		}
	}`

	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	pkg, err := LoadPackageJSON(path)
	require.NoError(t, err)

	pkg.setDependency("dependencies", "alpha", "^9.0.0")

	require.NoError(t, SavePackageJSON(path, pkg))

	loaded, err := LoadPackageJSON(path)
	require.NoError(t, err)

	assert.Equal(t, "^9.0.0", loaded.Dependencies.Values["alpha"])
	assert.Equal(t, []string{"alpha", "zebra"}, loaded.Dependencies.Keys)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
