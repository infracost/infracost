package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ECR struct {
	Address   *string
	Region    *string
	StorageGb *float64 `infracost_usage:"storage_gb"`
}

var ECRUsageSchema = []*schema.UsageItem{{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *ECR) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ECR) BuildResource() *schema.Resource {
	region := *r.Region

	var storageSize *decimal.Decimal

	if r != nil && r.StorageGb != nil {
		storageSize = decimalPtr(decimal.NewFromFloat(*r.StorageGb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: storageSize,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonECR"),
					ProductFamily: strPtr("EC2 Container Registry"),
				},
			},
		}, UsageSchema: ECRUsageSchema,
	}
}
