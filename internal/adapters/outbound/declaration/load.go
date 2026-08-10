package declaration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/goccy/go-yaml"
	declarationport "github.com/jamessawle/sbxflow/internal/ports/declaration"
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
func Load(data []byte) (declarationport.Configuration, error) {
	document, jsonData, err := decodeDocument(data)
	if err != nil {
		return declarationport.Configuration{}, err
	}

	schema, err := configurationSchema()
	if err != nil {
		return declarationport.Configuration{}, err
	}
	if err := schema.Validate(document); err != nil {
		return declarationport.Configuration{}, fmt.Errorf("configuration schema validation failed: %w", err)
	}

	var documentConfiguration declarationport.Configuration
	if err := json.Unmarshal(jsonData, &documentConfiguration); err != nil {
		return declarationport.Configuration{}, fmt.Errorf("decode validated configuration: %w", err)
	}
	return documentConfiguration, nil
}

// LoadLifecycleTarget parses only the supported declaration version and
// sandbox name needed by teardown lifecycle operations.
func LoadLifecycleTarget(data []byte) (declarationport.LifecycleTarget, error) {
	document, _, err := decodeDocument(data)
	if err != nil {
		return declarationport.LifecycleTarget{}, err
	}

	root, ok := document.(map[string]any)
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity must be an object")
	}
	version, ok := root["version"]
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity is missing version")
	}
	versionNumber, ok := version.(float64)
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity version must be an integer")
	}
	if versionNumber != 1 {
		return declarationport.LifecycleTarget{}, fmt.Errorf("unsupported configuration version %v", version)
	}

	sandbox, ok := root["sandbox"]
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity is missing sandbox")
	}
	sandboxObject, ok := sandbox.(map[string]any)
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity sandbox must be an object")
	}
	name, ok := sandboxObject["name"]
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity is missing sandbox.name")
	}
	nameString, ok := name.(string)
	if !ok {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity sandbox.name must be a string")
	}
	if nameString == "" {
		return declarationport.LifecycleTarget{}, errors.New("configuration identity sandbox.name must not be empty")
	}
	return declarationport.LifecycleTarget{Name: nameString}, nil
}

func decodeDocument(data []byte) (any, []byte, error) {
	var raw any
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil, errors.New("configuration is empty")
		}
		return nil, nil, fmt.Errorf("parse YAML: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, nil, fmt.Errorf("parse YAML: %w", err)
		}
		return nil, nil, errors.New("configuration must contain exactly one YAML document")
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("convert YAML to JSON: %w", err)
	}

	var document any
	if err := json.Unmarshal(jsonData, &document); err != nil {
		return nil, nil, fmt.Errorf("convert YAML to JSON: %w", err)
	}
	return document, jsonData, nil
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
