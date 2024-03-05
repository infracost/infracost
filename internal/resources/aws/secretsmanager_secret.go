package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SecretsManagerSecret struct {
	Address         string
	Region          string
	MonthlyRequests *int64 `infracost_usage:"monthly_requests"`
}

func (r *SecretsManagerSecret) CoreType() string {
	return "SecretsManagerSecret"
}

func (r *SecretsManagerSecret) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *SecretsManagerSecret) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SecretsManagerSecret) BuildResource() *schema.Resource {
	var monthlyRequests *decimal.Decimal
	if r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Secret",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
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
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSSecretsManager"),
					ProductFamily: strPtr("API Request"),
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
