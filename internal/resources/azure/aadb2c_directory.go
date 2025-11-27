package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type AADB2CDirectory struct {
	Address            string
	Region             string
	MonthlyActiveUsers *int64 `infracost_usage:"monthly_active_users"`
}

func (r *AADB2CDirectory) CoreType() string {
	return "AADB2CDirectory"
}

func (r *AADB2CDirectory) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "monthly_active_users",
			DefaultValue: 0,
			ValueType:    schema.Int64,
		},
	}
}

func (r *AADB2CDirectory) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AADB2CDirectory) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	freeTier := int64(50000)
	billableMAU := int64(0)
	if r.MonthlyActiveUsers != nil && *r.MonthlyActiveUsers > freeTier {
		billableMAU = *r.MonthlyActiveUsers - freeTier
	}

	if billableMAU > 0 {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Monthly active users (over free tier)",
			Unit:           "users",
			UnitMultiplier: decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(billableMAU)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Azure Active Directory B2C"),
				ProductFamily: strPtr("Security"),
				AttributeFilters: []*schema.AttributeFilter{
					// TODO: Confirm the correct meterName for B2C MAU pricing
					{Key: "meterName", Value: strPtr("Monthly Active Users")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
} 