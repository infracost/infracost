package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

func GetNewKMSExternalKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_external_key",
		RFunc: NewKMSExternalKey,
	}
}

func NewKMSExternalKey(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{
		CustomerMasterKeyCostComponent(region),
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
