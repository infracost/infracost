package main

import (
	"bytes"
	"github.com/pmezard/go-difflib/difflib"
	"os"
	"testing"
)

var schemaFile = "../../schema/infracost.schema.json"

func TestVerifyExample(t *testing.T) {
	generatedBytes, err := generateJSONSchema()
	if err != nil {
		t.Fatal(err)
	}

	exampleBytes, err := os.ReadFile(schemaFile)
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
		t.Fatalf("\nGenerated JSON schema does not match example.  Run `make jsonschema` to update:: \n\n%s\n", diff)
	}
}
