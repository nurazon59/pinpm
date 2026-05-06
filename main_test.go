package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	bin := filepath.Join(dir, "pinpm")

	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	return bin
}

func TestVersion(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "--version")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "v0.1.0\n", string(out))
}

func TestHelp(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "--help")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	require.True(t, bytes.Contains(out, []byte("Usage: pinpm <command>")))
	require.True(t, bytes.Contains(out, []byte("--config=STRING")))
	require.True(t, bytes.Contains(out, []byte("Commands:")))
}

func setupMockRegistry(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/express":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"name": "express",
				"dist-tags": {"latest": "4.21.0"},
				"versions": {
					"4.0.0": {"version": "4.0.0"},
					"4.10.0": {"version": "4.10.0"},
					"4.21.0": {"version": "4.21.0"}
				}
			}`))
		case "/lodash":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"name": "lodash",
				"dist-tags": {"latest": "4.17.21"},
				"versions": {
					"4.17.0": {"version": "4.17.0"},
					"4.17.15": {"version": "4.17.15"},
					"4.17.21": {"version": "4.17.21"}
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(server.Close)
	return server
}

func TestPinCommandE2E(t *testing.T) {
	bin := buildBinary(t)
	server := setupMockRegistry(t)

	src := filepath.Join("testdata", "package.json")
	dst := filepath.Join(t.TempDir(), "package.json")

	data, err := os.ReadFile(src)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dst, data, 0644))

	cmd := exec.Command(bin, "pin", "-f", dst)
	cmd.Env = append(os.Environ(), "npm_config_registry="+server.URL)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	require.Contains(t, string(out), "Pinned 2 dependencies")

	result, err := os.ReadFile(dst)
	require.NoError(t, err)
	require.Contains(t, string(result), `"express": "4.21.0"`)
	require.Contains(t, string(result), `"lodash": "4.17.21"`)
}

func TestCheckCommandE2E(t *testing.T) {
	bin := buildBinary(t)
	server := setupMockRegistry(t)

	src := filepath.Join("testdata", "package.json")
	dst := filepath.Join(t.TempDir(), "package.json")

	data, err := os.ReadFile(src)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dst, data, 0644))

	cmd := exec.Command(bin, "check", "-f", dst)
	cmd.Env = append(os.Environ(), "npm_config_registry="+server.URL)
	out, _ := cmd.CombinedOutput()

	require.Equal(t, 1, cmd.ProcessState.ExitCode())
	require.Contains(t, string(out), "express (dependencies): ^4.0.0 -> 4.21.0")
	require.Contains(t, string(out), "lodash (devDependencies): ~4.17.0 -> 4.17.21")
}
