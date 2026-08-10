package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/goccy/go-yaml"
	publicschema "github.com/jamessawle/sbxflow/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaURL = "https://github.com/jamessawle/sbxflow/blob/main/schema/sbxflow.schema.json"

var (
	compiledSchema     *jsonschema.Schema
	compiledSchemaErr  error
	compiledSchemaOnce sync.Once
)

// Load parses exactly one YAML document, validates it against the published
// schema, and returns its typed representation.
func Load(data []byte) (Configuration, error) {
	var raw any
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return Configuration{}, errors.New("configuration is empty")
		}
		return Configuration{}, fmt.Errorf("parse YAML: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return Configuration{}, fmt.Errorf("parse YAML: %w", err)
		}
		return Configuration{}, errors.New("configuration must contain exactly one YAML document")
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		return Configuration{}, fmt.Errorf("convert YAML to JSON: %w", err)
	}

	var document any
	if err := json.Unmarshal(jsonData, &document); err != nil {
		return Configuration{}, fmt.Errorf("convert YAML to JSON: %w", err)
	}

	schema, err := configurationSchema()
	if err != nil {
		return Configuration{}, err
	}
	if err := schema.Validate(document); err != nil {
		return Configuration{}, fmt.Errorf("configuration schema validation failed: %w", err)
	}

	var configuration Configuration
	if err := json.Unmarshal(jsonData, &configuration); err != nil {
		return Configuration{}, fmt.Errorf("decode validated configuration: %w", err)
	}
	return configuration, nil
}

func configurationSchema() (*jsonschema.Schema, error) {
	compiledSchemaOnce.Do(func() {
		var document any
		if err := json.Unmarshal(publicschema.Configuration, &document); err != nil {
			compiledSchemaErr = fmt.Errorf("decode embedded configuration schema: %w", err)
			return
		}

		compiler := jsonschema.NewCompiler()
		compiler.DefaultDraft(jsonschema.Draft2020)
		if err := compiler.AddResource(schemaURL, document); err != nil {
			compiledSchemaErr = fmt.Errorf("load embedded configuration schema: %w", err)
			return
		}
		compiledSchema, compiledSchemaErr = compiler.Compile(schemaURL)
		if compiledSchemaErr != nil {
			compiledSchemaErr = fmt.Errorf("compile embedded configuration schema: %w", compiledSchemaErr)
		}
	})
	return compiledSchema, compiledSchemaErr
}
