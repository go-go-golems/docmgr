package skills

import (
	"bytes"
	"io"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// LoadPlan reads and validates a skill.yaml plan.
func LoadPlan(path string) (*Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read skill plan")
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var plan Plan
	if err := decoder.Decode(&plan); err != nil {
		return nil, errors.Wrap(err, "failed to decode skill plan")
	}

	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("skill plan must contain a single YAML document")
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return &plan, nil
}
