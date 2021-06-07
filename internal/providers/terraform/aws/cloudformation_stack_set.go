package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetCloudFormationStackSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudformation_stack_set",
		RFunc: NewCloudFormationStackSet,
	}
}

func NewCloudFormationStackSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	if d.Get("template_body").Type != gjson.Null && (checkAWS(d) || checkAlexa(d) || checkCustom(d)) {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cloudFormationCostComponents(d, u),
	}
}
