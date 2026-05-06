package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	fixedVersionRe = regexp.MustCompile(`^\d+\.\d+\.\d+(-[\w.]+)?(\+[\w.]+)?$`)
	wildcardRe     = regexp.MustCompile(`^\d+\.x(\.x)?$`)
)

func IsPinned(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	return fixedVersionRe.MatchString(v)
}

func ShouldSkip(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return true
	}
	skipPrefixes := []string{
		"file:",
		"git:",
		"git+",
		"github:",
		"workspace:",
		"npm:",
		"http:",
		"https:",
	}
	for _, p := range skipPrefixes {
		if strings.HasPrefix(v, p) {
			return true
		}
	}
	return false
}

func ResolveVersion(versionRange string, availableVersions []string) (string, error) {
	versionRange = strings.TrimSpace(versionRange)

	if versionRange == "*" || versionRange == "latest" {
		return latestStable(availableVersions)
	}

	if wildcardRe.MatchString(versionRange) {
		major := strings.Split(versionRange, ".")[0]
		nextMajor := func(m string) string {
			n := 0
			fmt.Sscanf(m, "%d", &n)
			return fmt.Sprintf("%d.0.0", n+1)
		}
		constraints, err := semver.NewConstraint(">= " + major + ".0.0, < " + nextMajor(major))
		if err != nil {
			return latestStable(availableVersions)
		}
		return resolveWithConstraint(constraints, availableVersions)
	}

	constraints, err := semver.NewConstraint(versionRange)
	if err != nil {
		return latestStable(availableVersions)
	}

	return resolveWithConstraint(constraints, availableVersions)
}

func resolveWithConstraint(constraints *semver.Constraints, versions []string) (string, error) {
	var matching []*semver.Version
	for _, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		if constraints.Check(sv) {
			matching = append(matching, sv)
		}
	}

	if len(matching) == 0 {
		return latestStable(versions)
	}

	sort.Sort(semver.Collection(matching))
	return matching[len(matching)-1].Original(), nil
}

func latestStable(versions []string) (string, error) {
	var stable []*semver.Version
	for _, v := range versions {
		sv, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		if sv.Prerelease() == "" {
			stable = append(stable, sv)
		}
	}

	if len(stable) == 0 {
		return "", nil
	}

	sort.Sort(semver.Collection(stable))
	return stable[len(stable)-1].Original(), nil
}
