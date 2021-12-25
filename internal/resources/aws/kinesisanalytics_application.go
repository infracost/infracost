package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type KinesisAnalyticsApplication struct {
	Address                *string
	Region                 *string
	KinesisProcessingUnits *int64 `infracost_usage:"kinesis_processing_units"`
}

var KinesisAnalyticsApplicationUsageSchema = []*schema.UsageItem{{Key: "kinesis_processing_units", ValueType: schema.Int64, DefaultValue: 0}}

func (r *KinesisAnalyticsApplication) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KinesisAnalyticsApplication) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)
	var kinesisProcessingUnits *decimal.Decimal

	if r != nil && r.KinesisProcessingUnits != nil {
		kinesisProcessingUnits = decimalPtr(decimal.NewFromInt(*r.KinesisProcessingUnits))
	}
	costComponents = append(costComponents, kinesisProcessingsCostComponent("Processing (stream)", region, kinesisProcessingUnits))
	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: KinesisAnalyticsApplicationUsageSchema,
	}
}
