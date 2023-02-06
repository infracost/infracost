package azurerm

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
)

type ResourceTests struct {
	expected *schema.Resource
}

func TestParseWhatifJson(t *testing.T) {
	// tests := []ResourceTests{
	// 	{
	// 		expected: &schema.Resource{
	// 			Name:           "my-resource-group2",
	// 			ResourceType:   "Microsoft.Resources/resourceGroups",
	// 			IsSkipped:      true,
	// 			NoPrice:        true,
	// 			CostComponents: []*schema.CostComponent{},
	// 		},
	// 	},
	// }

	// testFile, err := ioutil.ReadFile("./testdata/whatif-single.json")
	// if err != nil {
	// 	log.Fatal("Error reading test whatif", err)
	// }

	// parsed := gjson.ParseBytes(testFile)

	// ctx := config.NewProjectContext(config.EmptyRunContext(), &config.Project{}, log.Fields{})
	// parser := NewParser(ctx)

	// partials := parser.parseWhatifJSONResources(false, nil, nil, parsed)

	// for _, tests := range tests {
	// 	var resource *schema.Resource

	// }

	// testJson := json.Marshal()
}
