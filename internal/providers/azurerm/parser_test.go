package azurerm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type ResourceTests struct {
	expected *schema.Resource
}

func TestParseWhatifJson(t *testing.T) {
	tests := []ResourceTests{
		{
			expected: &schema.Resource{
				Name:         "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group2",
				ResourceType: "azurerm_resource_group",
				IsSkipped:    true,
				NoPrice:      true,
			},
		},
	}

	testFile, err := os.ReadFile("./testdata/whatif-single.json")
	if err != nil {
		log.Fatal("Error reading test whatif", err)
	}

	var whatif WhatIf
	err = json.Unmarshal(testFile, &whatif)
	if err != nil {
		log.Fatal("Error reading test whatif", err)
	}

	usage := schema.NewEmptyUsageMap()
	ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})
	parser := NewParser(ctx)

	partials, err := parser.parseResources(false, &whatif, usage)
	if err != nil {
		log.Fatal("Failed to create partial resources", err)
	}

	actual := make([]*schema.Resource, len(partials))
	for i, partial := range partials {
		actual[i] = schema.BuildResource(partial, nil)
	}

	for _, test := range tests {
		var resource *schema.Resource
		for _, act := range actual {
			if test.expected.Name == act.Name {
				resource = act
			}
		}

		assert.NotNil(t, resource)
		assert.Equal(t, test.expected.Name, resource.Name)
		assert.Equal(t, test.expected.ResourceType, resource.ResourceType)
		assert.Equal(t, test.expected.IsSkipped, resource.IsSkipped)
		assert.Equal(t, test.expected.NoPrice, resource.NoPrice)

		if test.expected.CostComponents != nil {
			assert.Equal(t, len(test.expected.CostComponents), len(resource.CostComponents))

			for i, expcc := range test.expected.CostComponents {
				actcc := resource.CostComponents[i]

				if expcc.MonthlyQuantity != nil {
					assert.Equal(t, expcc.MonthlyQuantity.BigInt(), actcc.MonthlyQuantity.BigInt())
				} else {
					assert.Equal(t, expcc.MonthlyQuantity, actcc.MonthlyQuantity)
				}
			}

		} else {
			assert.Equal(t, test.expected.CostComponents, resource.CostComponents)
		}
	}
}
