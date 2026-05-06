package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PackageJSON struct {
	Name                 string            `json:"name,omitempty"`
	Version              string            `json:"version,omitempty"`
	Dependencies         map[string]string `json:"dependencies,omitempty"`
	DevDependencies      map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies     map[string]string `json:"peerDependencies,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
}

type Dependency struct {
	Name    string
	Version string
	Section string
}

func (p *PackageJSON) AllDependencies() []Dependency {
	var deps []Dependency

	for section, m := range map[string]map[string]string{
		"dependencies":         p.Dependencies,
		"devDependencies":      p.DevDependencies,
		"peerDependencies":     p.PeerDependencies,
		"optionalDependencies": p.OptionalDependencies,
	} {
		if m == nil {
			continue
		}
		for name, ver := range m {
			deps = append(deps, Dependency{
				Name:    name,
				Version: ver,
				Section: section,
			})
		}
	}

	return deps
}

func (p *PackageJSON) setDependency(section, name, version string) error {
	var target *map[string]string
	switch section {
	case "dependencies":
		target = &p.Dependencies
	case "devDependencies":
		target = &p.DevDependencies
	case "peerDependencies":
		target = &p.PeerDependencies
	case "optionalDependencies":
		target = &p.OptionalDependencies
	default:
		return fmt.Errorf("unknown dependency section: %s", section)
	}
	if *target == nil {
		*target = make(map[string]string)
	}
	(*target)[name] = version
	return nil
}

func LoadPackageJSON(path string) (*PackageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	return &pkg, nil
}

func SavePackageJSON(path string, pkg *PackageJSON) error {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
