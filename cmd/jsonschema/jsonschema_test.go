package main

import (
	"bytes"
	"github.com/pmezard/go-difflib/difflib"
	"os"
	"testing"
)

var outputSchemaFile = "../../schema/infracost.schema.json"
var configSchemaFile = "../../schema/config.schema.json"

func TestVerifyOutputExample(t *testing.T) {
	generatedBytes, err := generateOutputJSONSchema()
	if err != nil {
		t.Fatal(err)
	}

	exampleBytes, err := os.ReadFile(outputSchemaFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(generatedBytes, exampleBytes) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(generatedBytes)),
			B:        difflib.SplitLines(string(exampleBytes)),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})
		t.Fatalf("\nGenerated output file JSON schema does not match example.  Run `make jsonschema` to update:: \n\n%s\n", diff)
	}
}

func TestVerifyConfigExample(t *testing.T) {
	generatedBytes, err := generateConfigFileJSONSchema()
	if err != nil {
		t.Fatal(err)
	}

	exampleBytes, err := os.ReadFile(configSchemaFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(generatedBytes, exampleBytes) {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(generatedBytes)),
			B:        difflib.SplitLines(string(exampleBytes)),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})
		t.Fatalf("\nGenerated config file JSON schema does not match example.  Run `make jsonschema` to update:: \n\n%s\n", diff)
	}
}
