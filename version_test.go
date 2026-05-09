package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPinned(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"fixed":            {"1.2.3", true},
		"fixed prerelease": {"1.2.3-beta.1", true},
		"fixed build":      {"1.2.3+build.123", true},
		"caret":            {"^1.2.3", false},
		"tilde":            {"~1.2.3", false},
		"gte":              {">=1.0.0", false},
		"wildcard":         {"1.x", false},
		"star":             {"*", false},
		"latest":           {"latest", false},
		"empty":            {"", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsPinned(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldSkip(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"file":      {"file:../pkg", true},
		"git":       {"git+https://github.com/foo/bar.git", true},
		"github":    {"github:user/repo", true},
		"workspace": {"workspace:*", true},
		"http":      {"http://example.com/pkg.tgz", true},
		"https":     {"https://example.com/pkg.tgz", true},
		"npm alias": {"npm:old-pkg@1.0.0", true},
		"fixed":     {"1.2.3", false},
		"caret":     {"^1.2.3", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ShouldSkip(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveVersionUnit(t *testing.T) {
	versions := []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0", "2.1.0", "3.0.0"}

	tests := map[string]struct {
		rangeSpec string
		want      string
	}{
		"caret major":    {"^1.0.0", "1.2.0"},
		"caret minor":    {"^1.1.0", "1.2.0"},
		"tilde minor":    {"~1.1.0", "1.1.0"},
		"tilde patch":    {"~1.2.0", "1.2.0"},
		"gte":            {">=2.0.0", "3.0.0"},
		"gt":             {">2.0.0", "3.0.0"},
		"wildcard":       {"*", "3.0.0"},
		"latest":         {"latest", "3.0.0"},
		"major wildcard": {"1.x", "1.2.0"},
		"no match":       {">=4.0.0", "3.0.0"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ResolveVersion(tt.rangeSpec, versions)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
