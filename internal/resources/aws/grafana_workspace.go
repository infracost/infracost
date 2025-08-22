package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type GrafanaWorkspace struct {
	Address                      string
	Region                       string
	License                      string
	EditorsAdministratorLicenses *int64 `infracost_usage:"editors_administrator_licenses"`
	ViewerLicenses               *int64 `infracost_usage:"viewer_licenses"`
}

func (r *GrafanaWorkspace) CoreType() string {
	return "GrafanaWorkspace"
}

func (r *GrafanaWorkspace) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "editors_administrator_licenses", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "viewer_licenses", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *GrafanaWorkspace) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *GrafanaWorkspace) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	var editorLicenses *decimal.Decimal
	if r.EditorsAdministratorLicenses != nil && *r.EditorsAdministratorLicenses > 0 {
		editorLicenses = decimalPtr(decimal.NewFromInt(*r.EditorsAdministratorLicenses))
	} else if r.EditorsAdministratorLicenses == nil {
		editorLicenses = decimalPtr(decimal.NewFromInt(1))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Editor/administrator licenses",
		Unit:            "users",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: editorLicenses,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonGrafana"),
			ProductFamily: strPtr("User Fees"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("Grafana:EditorUser$")},
			},
		},
	})

	var viewerLicenses *decimal.Decimal
	if r.ViewerLicenses != nil && *r.ViewerLicenses > 0 {
		viewerLicenses = decimalPtr(decimal.NewFromInt(*r.ViewerLicenses))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Viewer licenses",
		Unit:            "users",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: viewerLicenses,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonGrafana"),
			ProductFamily: strPtr("User Fees"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("Grafana:ViewerUser$")},
			},
		},
		UsageBased: true,
	})

	if strings.EqualFold(r.License, "ENTERPRISE") {
		var enterprisePluginsQty decimal.Decimal
		if editorLicenses != nil {
			enterprisePluginsQty = enterprisePluginsQty.Add(*editorLicenses)
		}
		if viewerLicenses != nil {
			enterprisePluginsQty = enterprisePluginsQty.Add(*viewerLicenses)
		}

		if enterprisePluginsQty.GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Enterprise plugins licenses",
				Unit:            "users",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: &enterprisePluginsQty,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonGrafana"),
					ProductFamily: strPtr("User Fees"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: regexPtr("Grafana:EnterprisePluginsUser$")},
					},
				},
			})
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
