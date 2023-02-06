package azurerm

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"path"
	"testing"
)

func TestWhatifFiles(t *testing.T) {
	testDataPath := "./testdata"
	testFiles := []string{"whatif-single.json"}

	for _, f := range testFiles {

		var whatIf WhatIf
		filePath := path.Join(testDataPath, f)

		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read test json file")
		}

		json.Unmarshal(file, &whatIf)

		after := whatIf.Properties.Changes[0].MarshalAfter()

		assert.Equal(t, after.Get("type"), "Microsoft.Resources/resourceGroups")
	}
}
