package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetSecretsManagerSecret() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_secretsmanager_secret",
		RFunc: NewSecretsManagerSecret,
	}
}

func NewSecretsManagerSecret(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monthlyRequests *decimal.Decimal

	if u != nil && u.Get("monthly_requests").Exists() {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Secret",
				Unit:            "months",
				UnitMultiplier:  1,
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
				Unit:            "requests",
				UnitMultiplier:  10000,
				MonthlyQuantity: monthlyRequests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSSecretsManager"),
					ProductFamily: strPtr("API Request"),
				},
			},
		},
	}
}
