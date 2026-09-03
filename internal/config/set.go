package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Set updates one dotted path (e.g. "theme.name") in the config file to a
// scalar value, creating the file and any missing intermediate maps. It edits
// the YAML node tree rather than re-marshalling a Config, so user comments,
// ordering, and unrelated keys survive.
func Set(path string, value any) error {
	file, err := Path()
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var doc yaml.Node
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			return fmt.Errorf("parsing %s: %w", file, err)
		}
	}
	if doc.Kind == 0 || len(doc.Content) == 0 {
		doc = yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}},
		}
	}

	if err := setPath(doc.Content[0], strings.Split(path, "."), value); err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return err
	}
	_ = enc.Close()

	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	mode := os.FileMode(0o600) // the file may hold a Linear token
	if fi, err := os.Stat(file); err == nil {
		mode = fi.Mode().Perm()
	}
	return os.WriteFile(file, buf.Bytes(), mode)
}

// setPath walks/creates mapping entries down segs and encodes value into the
// leaf node, keeping any comments attached to existing nodes.
func setPath(n *yaml.Node, segs []string, value any) error {
	for i, seg := range segs {
		last := i == len(segs)-1
		child := mappingValue(n, seg)
		if child == nil {
			key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: seg}
			child = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			n.Content = append(n.Content, key, child)
		}
		if last {
			return encodeInto(child, value)
		}
		if child.Kind != yaml.MappingNode {
			// A scalar sits where we need a map (e.g. `linear: ""`): replace it.
			child.Kind, child.Tag, child.Value, child.Content = yaml.MappingNode, "!!map", "", nil
		}
		n = child
	}
	return nil
}

// mappingValue returns the value node for key in a mapping, or nil.
func mappingValue(n *yaml.Node, key string) *yaml.Node {
	if n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

// encodeInto re-types an existing node to hold value, preserving its comments.
func encodeInto(n *yaml.Node, value any) error {
	var tmp yaml.Node
	if err := tmp.Encode(value); err != nil {
		return err
	}
	n.Kind, n.Style, n.Tag, n.Value, n.Content = tmp.Kind, tmp.Style, tmp.Tag, tmp.Value, tmp.Content
	return nil
}
