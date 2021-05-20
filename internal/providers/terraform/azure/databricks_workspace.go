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
		RFunc: NewAzureDatabricksWorkspace,
	}
}

func NewAzureDatabricksWorkspace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	location := d.Get("location").String()
	sku := strings.Title(d.Get("sku").String())

	if sku == "Trial" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var allPurpose, jobs, jobsLight *decimal.Decimal

	if u != nil && u.Get("monthly_all_purpose_compute_dbu").Exists() {
		allPurpose = decimalPtr(decimal.NewFromInt(u.Get("monthly_all_purpose_compute_dbu").Int()))
	}
	costComponents = append(costComponents, databricksCostComponent(
		"All-purpose compute DBUs",
		location,
		fmt.Sprintf("%s All-purpose Compute", sku),
		allPurpose,
	))

	if u != nil && u.Get("monthly_jobs_compute_dbu").Exists() {
		jobs = decimalPtr(decimal.NewFromInt(u.Get("monthly_jobs_compute_dbu").Int()))
	}
	costComponents = append(costComponents, databricksCostComponent(
		"Jobs compute DBUs",
		location,
		fmt.Sprintf("%s Jobs Compute", sku),
		jobs,
	))

	if u != nil && u.Get("monthly_jobs_light_compute_dbu").Exists() {
		jobsLight = decimalPtr(decimal.NewFromInt(u.Get("monthly_jobs_light_compute_dbu").Int()))
	}
	costComponents = append(costComponents, databricksCostComponent(
		"Jobs light compute DBUs",
		location,
		fmt.Sprintf("%s Jobs Light Compute", sku),
		jobsLight,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func databricksCostComponent(name, location, skuName string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "DBU-hours",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
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
