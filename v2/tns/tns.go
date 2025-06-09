package tns

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Parse parses a TNS string and returns a TNS struct.
func Parse(tnsStr string) (*TNS, error) {
	yamlData, err := toYaml(tnsStr)
	if err != nil {
		return nil, err
	}
	var tns TNS
	err = yaml.Unmarshal(yamlData, &tns)
	if err != nil {
		return nil, err
	}
	return &tns, nil
}

func toYaml(tns string) ([]byte, error) {
	p := parser{
		tns:           []rune(tns),
		lowerCaseKeys: true,
	}
	parsedTns, err := p.consume()
	if err != nil {
		return nil, fmt.Errorf("failed to parse TNS: %w", err)
	}
	tnsYaml, err := yaml.Marshal(parsedTns)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return tnsYaml, nil
}
