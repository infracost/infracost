package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetNewKMSExternalKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_external_key",
		RFunc: NewKMSExternalKey,
	}
}

func NewKMSExternalKey(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{
		CustomerMasterKeyCostComponent(region),
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
