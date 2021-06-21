package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKenesisAnalyticsApplicationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesis_analytics_application",
		RFunc: NewKenesisAnalyticsApplication,
		Notes: []string{
			"Terraform doesn’t currently support Analytics Studio, but when it does they will require 2 orchestration KPUs.",
		},
	}
}
func NewKenesisAnalyticsApplication(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	var kinesisProcessingUnits *decimal.Decimal

	if u != nil && u.Get("kinesis_processing_units").Type != gjson.Null {
		kinesisProcessingUnits = decimalPtr(decimal.NewFromInt(u.Get("kinesis_processing_units").Int()))
	}
	costComponents = append(costComponents, kenesisProcessingsCostComponent("Processing (stream)", region, kinesisProcessingUnits))
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
