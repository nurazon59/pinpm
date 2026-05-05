package pinpm

import (
	"fmt"
)

type Result struct {
	Pinned  []Change
	Skipped []Skip
}

type Change struct {
	Name       string
	Section    string
	OldVersion string
	NewVersion string
}

type Skip struct {
	Name    string
	Section string
	Reason  string
}

type Checker struct {
	client *Client
}

func NewChecker(client *Client) *Checker {
	return &Checker{client: client}
}

func (c *Checker) Check(pkg *PackageJSON) (*Result, error) {
	result := &Result{}

	for _, dep := range pkg.AllDependencies() {
		if ShouldSkip(dep.Version) {
			result.Skipped = append(result.Skipped, Skip{
				Name:    dep.Name,
				Section: dep.Section,
				Reason:  fmt.Sprintf("skippable version: %s", dep.Version),
			})
			continue
		}

		if IsPinned(dep.Version) {
			result.Skipped = append(result.Skipped, Skip{
				Name:    dep.Name,
				Section: dep.Section,
				Reason:  "already pinned",
			})
			continue
		}

		resolved, err := c.client.ResolveVersion(dep.Name, dep.Version)
		if err != nil {
			return nil, fmt.Errorf("resolve version for %s: %w", dep.Name, err)
		}

		result.Pinned = append(result.Pinned, Change{
			Name:       dep.Name,
			Section:    dep.Section,
			OldVersion: dep.Version,
			NewVersion: resolved,
		})
	}

	return result, nil
}

func (c *Checker) Apply(path string, pkg *PackageJSON, result *Result) error {
	for _, change := range result.Pinned {
		if err := pkg.setDependency(change.Section, change.Name, change.NewVersion); err != nil {
			return fmt.Errorf("update %s: %w", change.Name, err)
		}
	}

	if err := SavePackageJSON(path, pkg); err != nil {
		return fmt.Errorf("save package.json: %w", err)
	}

	return nil
}
