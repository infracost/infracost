package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/alecthomas/jsonschema"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/shopspring/decimal"
	"os"
	"reflect"
	"strings"
)

func main() {
	var c cmdConfig
	flag.StringVar(&c.Filename, "out-file", "", "The file to write with the generated JSON schema.")
	flag.StringVar(&c.Schema, "schema", "output", "The schema to write, 'output' (default) or 'config'")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if c.Filename == "" {
		exitWithErr(errors.New("Out file name cannot be blank"))
	}

	c.Filename = strings.ToLower(c.Filename)
	var b []byte
	var err error

	switch c.Schema {
	case "output":
		b, err = generateOutputJSONSchema()
	case "config":
		b, err = generateConfigFileJSONSchema()
	default:
		err = fmt.Errorf("Unknown schema, expected output or config:  %s", c.Schema)
	}

	if err != nil {
		exitWithErr(fmt.Errorf("Error generating JSON schema:\n%w", err))
	}

	err = writeOutput(c, b)
	if err != nil {
		exitWithErr(fmt.Errorf("Error writing file:\n%w", err))
	}
}

func typeMapper(i reflect.Type) *jsonschema.Type {
	if i == reflect.TypeFor[decimal.Decimal]() {
		return &jsonschema.Type{
			Type: "decimal",
		}
	}
	return nil
}

func generateOutputJSONSchema() ([]byte, error) {
	schemaReflector := &jsonschema.Reflector{
		TypeMapper: typeMapper,
	}

	schema := schemaReflector.Reflect(&output.Root{})

	// Recursive $refs cause Open Policy Agent to blow up, so tweak the Resource schema subresources to be non-recursive
	prop, ok := schema.Definitions["Resource"].Properties.Get("subresources")
	if !ok {
		return nil, fmt.Errorf("failed to find subresources property in Resource definition")
	}
	prop.(*jsonschema.Type).Items.Ref = "#/definitions/Subresource"

	// Add the type definition for subresources
	sub, err := subresourcesType()
	if err != nil {
		return nil, err
	}
	schema.Definitions["Subresource"] = sub

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	b = bytes.ReplaceAll(b, []byte("\"type\": \"decimal\""), []byte("\"type\": [\"string\", \"null\"]"))

	return b, nil
}

func subresourcesType() (*jsonschema.Type, error) {
	schemaReflector := &jsonschema.Reflector{
		TypeMapper:     typeMapper,
		ExpandedStruct: false,
	}

	subschema := schemaReflector.Reflect(&output.Resource{})

	// Avoid recursion by setting the subresources array of subresources to a generic object type
	subprop, ok := subschema.Definitions["Resource"].Properties.Get("subresources")
	if !ok {
		return nil, fmt.Errorf("failed to find subresources property in Resource definition")
	}
	subprop.(*jsonschema.Type).Items.Ref = ""
	subprop.(*jsonschema.Type).Items.Type = "object"

	return subschema.Definitions["Resource"], nil
}

func generateConfigFileJSONSchema() ([]byte, error) {
	schemaReflector := &jsonschema.Reflector{
		TypeMapper: typeMapper,
	}

	schema := schemaReflector.Reflect(&config.ConfigFileSpec{})

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func writeOutput(c cmdConfig, data []byte) error {
	return os.WriteFile(c.Filename, data, 0600)
}

type cmdConfig struct {
	Filename string
	Schema   string
}

func exitWithErr(err error) {
	fmt.Fprint(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}
