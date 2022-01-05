package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SecretsmanagerSecret struct {
	Address         *string
	Region          *string
	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

var SecretsmanagerSecretUsageSchema = []*schema.UsageItem{{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}}

func (r *SecretsmanagerSecret) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SecretsmanagerSecret) BuildResource() *schema.Resource {
	region := *r.Region

	var monthlyRequests *decimal.Decimal

	if r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Secret",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSSecretsManager"),
					ProductFamily: strPtr("Secret"),
				},
			},
			{
				Name:            "API requests",
				Unit:            "10k requests",
				UnitMultiplier:  decimal.NewFromInt(10000),
				MonthlyQuantity: monthlyRequests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSSecretsManager"),
					ProductFamily: strPtr("API Request"),
				},
			},
		}, UsageSchema: SecretsmanagerSecretUsageSchema,
	}
}
