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
	Extra                map[string]json.RawMessage
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

func (p *PackageJSON) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	knownFields := []string{"name", "version", "dependencies", "devDependencies", "peerDependencies", "optionalDependencies"}
	p.Extra = make(map[string]json.RawMessage)

	for key, val := range raw {
		isKnown := false
		for _, k := range knownFields {
			if key == k {
				isKnown = true
				break
			}
		}
		if isKnown {
			continue
		}
		p.Extra[key] = val
	}

	if v, ok := raw["name"]; ok {
		json.Unmarshal(v, &p.Name)
	}
	if v, ok := raw["version"]; ok {
		json.Unmarshal(v, &p.Version)
	}
	if v, ok := raw["dependencies"]; ok {
		json.Unmarshal(v, &p.Dependencies)
	}
	if v, ok := raw["devDependencies"]; ok {
		json.Unmarshal(v, &p.DevDependencies)
	}
	if v, ok := raw["peerDependencies"]; ok {
		json.Unmarshal(v, &p.PeerDependencies)
	}
	if v, ok := raw["optionalDependencies"]; ok {
		json.Unmarshal(v, &p.OptionalDependencies)
	}

	return nil
}

func (p PackageJSON) MarshalJSON() ([]byte, error) {
	raw := make(map[string]json.RawMessage)

	if p.Name != "" {
		v, _ := json.Marshal(p.Name)
		raw["name"] = v
	}
	if p.Version != "" {
		v, _ := json.Marshal(p.Version)
		raw["version"] = v
	}
	if p.Dependencies != nil {
		v, _ := json.Marshal(p.Dependencies)
		raw["dependencies"] = v
	}
	if p.DevDependencies != nil {
		v, _ := json.Marshal(p.DevDependencies)
		raw["devDependencies"] = v
	}
	if p.PeerDependencies != nil {
		v, _ := json.Marshal(p.PeerDependencies)
		raw["peerDependencies"] = v
	}
	if p.OptionalDependencies != nil {
		v, _ := json.Marshal(p.OptionalDependencies)
		raw["optionalDependencies"] = v
	}

	for key, val := range p.Extra {
		raw[key] = val
	}

	result, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}

func LoadPackageJSON(path string) (*PackageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var pkg PackageJSON
	if err := pkg.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	return &pkg, nil
}

func SavePackageJSON(path string, pkg *PackageJSON) error {
	data, err := pkg.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
