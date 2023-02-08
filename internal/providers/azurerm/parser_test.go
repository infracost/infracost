package azurerm

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/azurerm/util"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type ResourceTests struct {
	expected []*schema.Resource
	filePath string
}

func TestParseWhatifJson(t *testing.T) {
	tests := []ResourceTests{
		{
			filePath: "./testdata/what_if.json",
			expected: []*schema.Resource{
				{
					Name:         "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/rg",
					ResourceType: "azurerm_resource_group",
					IsSkipped:    true,
					NoPrice:      true,
					SkipMessage:  "Free resource",
				},
				{
					Name:         "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group/providers/Microsoft.Web/serverfarms/AppServicePlan-AzureLinuxApp",
					ResourceType: "azurerm_app_service_plan",
					IsSkipped:    false,
					NoPrice:      false,
					CostComponents: []*schema.CostComponent{
						{
							Name:           "Instance usage (S1)",
							HourlyQuantity: util.DecimalPtr(decimal.NewFromInt(1)),
						},
					},
				},
				{
					Name:         "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/my-resource-group/providers/Microsoft.Web/sites/AzureLinuxApp-webapp",
					ResourceType: "azurerm_windows_web_app",
					IsSkipped:    true,
					NoPrice:      true,
					SkipMessage:  "Free resource",
				},
			},
		},
	}

	for i, test := range tests {
		testFile, err := os.ReadFile(test.filePath)
		if err != nil {
			t.Fatalf("[Test %d]Error reading test whatif: "+err.Error(), i)
		}

		var whatif WhatIf
		err = json.Unmarshal(testFile, &whatif)
		if err != nil {
			t.Fatalf("[Test %d]Error reading test whatif: "+err.Error(), i)
		}

		usage := schema.NewEmptyUsageMap()
		ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})
		parser := NewParser(ctx)

		pastPartials, partials, err := parser.parse(testFile, usage)
		partials = append(partials, pastPartials...)

		if err != nil {
			t.Fatalf("[Test %d] Failed to create partial resources"+err.Error(), i)
		}

		actual := make([]*schema.Resource, len(partials))
		for i, partial := range partials {
			actual[i] = schema.BuildResource(partial, nil)
		}

		assert.Greater(t, len(partials), 0)

		var resource *schema.Resource
		for j, act := range actual {
			resIdMsg := fmt.Sprintf("Test: %d - Change: %d", i, j)

			expected := test.expected[j]
			if expected.Name == act.Name {
				resource = act
			}

			assert.NotNil(t, resource, resIdMsg)
			assert.Equal(t, expected.Name, resource.Name, resIdMsg)
			assert.Equal(t, expected.ResourceType, resource.ResourceType, resIdMsg)
			assert.Equal(t, expected.IsSkipped, resource.IsSkipped, resIdMsg)
			assert.Equal(t, expected.NoPrice, resource.NoPrice, resIdMsg)
			assert.Equal(t, expected.SkipMessage, resource.SkipMessage, resIdMsg)

			if expected.CostComponents != nil {
				assert.Equal(t, len(expected.CostComponents), len(resource.CostComponents))

				for i, expcc := range expected.CostComponents {
					actcc := resource.CostComponents[i]

					assert.Equal(t, expcc.Name, actcc.Name)

					if expcc.MonthlyQuantity != nil {
						assert.Equal(t, expcc.MonthlyQuantity.BigInt(), actcc.MonthlyQuantity.BigInt())
					} else {
						assert.Equal(t, expcc.MonthlyQuantity, actcc.MonthlyQuantity)
					}

					if expcc.HourlyQuantity != nil {
						assert.Equal(t, expcc.HourlyQuantity.BigInt(), actcc.HourlyQuantity.BigInt())
					} else {
						assert.Equal(t, expcc.HourlyQuantity, actcc.HourlyQuantity)
					}
				}

			} else {
				assert.Equal(t, expected.CostComponents, resource.CostComponents)
			}
		}
	}
}
