package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EcrRepository struct {
	Address   *string
	Region    *string
	StorageGb *float64 `infracost_usage:"storage_gb"`
}

var EcrRepositoryUsageSchema = []*schema.UsageItem{{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *EcrRepository) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EcrRepository) BuildResource() *schema.Resource {
	region := *r.Region

	var storageSize *decimal.Decimal

	if r.StorageGb != nil {
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
		}, UsageSchema: EcrRepositoryUsageSchema,
	}
}
