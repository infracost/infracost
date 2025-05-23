package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type GrafanaWorkspace struct {
	Address string
	Region  string
	License string
}

func (r *GrafanaWorkspace) CoreType() string {
	return "GrafanaWorkspace"
}

func (r *GrafanaWorkspace) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *GrafanaWorkspace) PopulateUsage(u *schema.UsageData) {
	// No usage data needed for Grafana workspace
}

func (r *GrafanaWorkspace) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	// Add cost component based on license type
	if r.License == "ENTERPRISE" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Enterprise license",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonGrafana"),
				ProductFamily: strPtr("Amazon Grafana"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "licenseModel", Value: strPtr("Enterprise")},
				},
			},
		})
	} else if r.License == "STANDARD" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Standard license",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonGrafana"),
				ProductFamily: strPtr("Amazon Grafana"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "licenseModel", Value: strPtr("Standard")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
	}
} 