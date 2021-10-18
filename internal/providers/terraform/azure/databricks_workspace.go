package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMDatabricksWorkspaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_databricks_workspace",
		RFunc: NewAzureRMDatabricksWorkspace,
	}
}

func NewAzureRMDatabricksWorkspace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	var costComponents []*schema.CostComponent

	sku := strings.Title(d.Get("sku").String())

	if sku == "Trial" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var allPurpose, jobs, jobsLight *decimal.Decimal

	if u != nil && u.Get("monthly_all_purpose_compute_dbu_hrs").Exists() {
		allPurpose = decimalPtr(decimal.NewFromInt(u.Get("monthly_all_purpose_compute_dbu_hrs").Int()))
	}
	costComponents = append(costComponents, databricksCostComponent(
		"All-purpose compute DBUs",
		region,
		fmt.Sprintf("%s All-purpose Compute", sku),
		allPurpose,
	))

	if u != nil && u.Get("monthly_jobs_compute_dbu_hrs").Exists() {
		jobs = decimalPtr(decimal.NewFromInt(u.Get("monthly_jobs_compute_dbu_hrs").Int()))
	}
	costComponents = append(costComponents, databricksCostComponent(
		"Jobs compute DBUs",
		region,
		fmt.Sprintf("%s Jobs Compute", sku),
		jobs,
	))

	if u != nil && u.Get("monthly_jobs_light_compute_dbu_hrs").Exists() {
		jobsLight = decimalPtr(decimal.NewFromInt(u.Get("monthly_jobs_light_compute_dbu_hrs").Int()))
	}
	costComponents = append(costComponents, databricksCostComponent(
		"Jobs light compute DBUs",
		region,
		fmt.Sprintf("%s Jobs Light Compute", sku),
		jobsLight,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func databricksCostComponent(name, region, skuName string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "DBU-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Databricks"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(skuName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
