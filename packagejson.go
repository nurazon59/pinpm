package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type OrderedMap struct {
	Keys   []string
	Values map[string]string
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys:   []string{},
		Values: make(map[string]string),
	}
}

func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	om.Keys = make([]string, 0, len(raw))
	om.Values = make(map[string]string)

	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if _, ok := t.(json.Delim); !ok || t.(json.Delim) != '{' {
		return fmt.Errorf("expected { got %v", t)
	}

	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("expected string key got %T", keyToken)
		}
		var val string
		if err := dec.Decode(&val); err != nil {
			return err
		}
		om.Keys = append(om.Keys, key)
		om.Values[key] = val
	}

	return nil
}

func (om OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range om.Keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		k, _ := json.Marshal(key)
		buf.Write(k)
		buf.WriteByte(':')
		v, _ := json.Marshal(om.Values[key])
		buf.Write(v)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (om *OrderedMap) Set(key, value string) {
	if _, exists := om.Values[key]; !exists {
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = value
}

type PackageJSON struct {
	Name                 string      `json:"name,omitempty"`
	Version              string      `json:"version,omitempty"`
	Dependencies         *OrderedMap `json:"dependencies,omitempty"`
	DevDependencies      *OrderedMap `json:"devDependencies,omitempty"`
	PeerDependencies     *OrderedMap `json:"peerDependencies,omitempty"`
	OptionalDependencies *OrderedMap `json:"optionalDependencies,omitempty"`
	Extra                map[string]json.RawMessage
}

type Dependency struct {
	Name    string
	Version string
	Section string
}

func (p *PackageJSON) AllDependencies() []Dependency {
	var deps []Dependency

	for section, om := range map[string]*OrderedMap{
		"dependencies":         p.Dependencies,
		"devDependencies":      p.DevDependencies,
		"peerDependencies":     p.PeerDependencies,
		"optionalDependencies": p.OptionalDependencies,
	} {
		if om == nil {
			continue
		}
		for _, name := range om.Keys {
			deps = append(deps, Dependency{
				Name:    name,
				Version: om.Values[name],
				Section: section,
			})
		}
	}

	return deps
}

func (p *PackageJSON) setDependency(section, name, version string) error {
	var target **OrderedMap
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
		*target = NewOrderedMap()
	}
	(*target).Set(name, version)
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
		v, _ := p.Dependencies.MarshalJSON()
		raw["dependencies"] = v
	}
	if p.DevDependencies != nil {
		v, _ := p.DevDependencies.MarshalJSON()
		raw["devDependencies"] = v
	}
	if p.PeerDependencies != nil {
		v, _ := p.PeerDependencies.MarshalJSON()
		raw["peerDependencies"] = v
	}
	if p.OptionalDependencies != nil {
		v, _ := p.OptionalDependencies.MarshalJSON()
		raw["optionalDependencies"] = v
	}

	for key, val := range p.Extra {
		raw[key] = val
	}

	result, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, err
	}

	// json.MarshalIndentは & < > をエスケープするため、元に戻す
	result = bytes.ReplaceAll(result, []byte(`\u0026`), []byte(`&`))
	result = bytes.ReplaceAll(result, []byte(`\u003c`), []byte(`<`))
	result = bytes.ReplaceAll(result, []byte(`\u003e`), []byte(`>`))

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
