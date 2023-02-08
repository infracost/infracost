package azurerm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests to check if serializing the Whatif JSON works properly, for development
// Pretty much testing Go's internal JSON system and gjson, so not particularly useful in the end
// TODO: Remove this when done with contribution
func TestWhatifSerialization(t *testing.T) {
	testDataPath := "./testdata"

	testFiles := []string{"whatif-single.json"}
	expected := []WhatIf{
		{
			Status: "Succeeded",
			Changes: []ResourceSnapshot{
				{
					ResourceId: "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group2",
					ChangeType: Create,
					AfterRaw:   []byte("{\"apiVersion\":\"2019-03-01\",\"id\":\"/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group2\",\"type\":\"Microsoft.Resources/resourceGroups\",\"name\":\"my-resource-group2\",\"location\":\"location3\"}"),
				},
			},
		},
	}

	for i, f := range testFiles {
		exp := expected[i]
		var whatIf WhatIf
		filePath := path.Join(testDataPath, f)

		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read test json file")
		}

		json.Unmarshal(file, &whatIf)

		assert.Equal(t, len(exp.Changes), len(whatIf.Changes))

		for j, c := range whatIf.Changes {
			wiBefore := c.Before()
			wiAfter := c.After()

			expBefore := exp.Changes[j].Before()
			expAfter := exp.Changes[j].After()

			assert.Equal(t, expBefore.Get("apiVersion").Str, wiBefore.Get("apiVersion").Str)
			assert.Equal(t, expBefore.Get("id").Str, wiBefore.Get("id").Str)
			assert.Equal(t, expBefore.Get("type").Str, wiBefore.Get("type").Str)

			assert.Equal(t, expAfter.Get("apiVersion").Str, wiAfter.Get("apiVersion").Str)
			assert.Equal(t, expAfter.Get("id").Str, wiAfter.Get("id").Str)
			assert.Equal(t, expAfter.Get("type").Str, wiAfter.Get("type").Str)
		}
	}
}
