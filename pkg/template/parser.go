// Package template provides functionality for loading and managing vulnerable target environment templates.
package template

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v3"
)

// loadTemplate loads a template from the specified directory by reading its index.yaml file.
func loadTemplate(filepath string) (Template, error) {
	var template Template
	file, err := os.ReadFile(path.Join(filepath, "index.yaml")) // #nosec: G304
	if err != nil {
		return template, err
	}
	if err = yaml.Unmarshal(file, &template); err != nil {
		return template, err
	}
	return template, template.Validate()
}

// loadPlaybookFromFile parses and validates a single playbook YAML file.
func loadPlaybookFromFile(filePath string) (Playbook, error) {
	var pb Playbook
	data, err := os.ReadFile(filePath) // #nosec: G304
	if err != nil {
		return pb, fmt.Errorf("failed to read playbook file: %w", err)
	}
	if err := yaml.Unmarshal(data, &pb); err != nil {
		return pb, fmt.Errorf("failed to parse playbook file: %w", err)
	}
	return pb, pb.Validate()
}
