package pinpm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const defaultRegistry = "https://registry.npmjs.org"

type Packument struct {
	Name     string            `json:"name"`
	DistTags map[string]string `json:"dist-tags"`
	Versions map[string]any    `json:"versions"`
}

type Client struct {
	registry string
}

func NewClient(registry string) *Client {
	if registry == "" {
		registry = DetectRegistry()
	}
	return &Client{
		registry: registry,
	}
}

func DetectRegistry() string {
	if env := os.Getenv("npm_config_registry"); env != "" {
		return env
	}

	if rc := parseNpmrc(); rc != "" {
		return rc
	}

	return defaultRegistry
}

func parseNpmrc() string {
	paths := []string{
		filepath.Join(".npmrc"),
		filepath.Join(os.Getenv("HOME"), ".npmrc"),
	}

	for _, path := range paths {
		if val := readNpmrc(path); val != "" {
			return val
		}
	}

	return ""
}

func readNpmrc(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "registry=") {
			return strings.TrimPrefix(line, "registry=")
		}
	}

	return ""
}

func (c *Client) GetPackage(name string) (*Packument, error) {
	url := fmt.Sprintf("%s/%s", c.registry, name)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, name)
	}

	var packument Packument
	if err := json.NewDecoder(resp.Body).Decode(&packument); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &packument, nil
}

func (c *Client) ResolveVersion(name, versionRange string) (string, error) {
	pkg, err := c.GetPackage(name)
	if err != nil {
		return "", err
	}

	versions := make([]string, 0, len(pkg.Versions))
	for v := range pkg.Versions {
		versions = append(versions, v)
	}

	return ResolveVersion(versionRange, versions)
}
