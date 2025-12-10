package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ECRRepository struct {
	Address   string
	Region    string
	StorageGB *float64 `infracost_usage:"storage_gb"`
}

func (r *ECRRepository) CoreType() string {
	return "ECRRepository"
}

func (r *ECRRepository) UsageSchema() []*schema.UsageItem {
	return ECRRepositoryUsageSchema
}

var ECRRepositoryUsageSchema = []*schema.UsageItem{
	{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *ECRRepository) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ECRRepository) BuildResource() *schema.Resource {
	var storageSize *decimal.Decimal
	if r.StorageGB != nil {
		storageSize = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: storageSize,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonECR"),
					ProductFamily: strPtr("EC2 Container Registry"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", Value: strPtr("TimedStorage-ByteHrs")},
					},
				},
				UsageBased: true,
			},
		},
		UsageSchema: ECRRepositoryUsageSchema,
	}
}
