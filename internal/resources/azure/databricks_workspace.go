package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type DatabricksWorkspace struct {
	Address                        string
	Region                         string
	SKU                            string
	MonthlyAllPurposeComputeDbuHrs *int64 `infracost_usage:"monthly_all_purpose_compute_dbu_hrs"`
	MonthlyJobsComputeDbuHrs       *int64 `infracost_usage:"monthly_jobs_compute_dbu_hrs"`
	MonthlyJobsLightComputeDbuHrs  *int64 `infracost_usage:"monthly_jobs_light_compute_dbu_hrs"`
}

var DatabricksWorkspaceUsageSchema = []*schema.UsageItem{{Key: "monthly_all_purpose_compute_dbu_hrs", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_jobs_compute_dbu_hrs", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_jobs_light_compute_dbu_hrs", ValueType: schema.Int64, DefaultValue: 0}}

func (r *DatabricksWorkspace) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DatabricksWorkspace) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	sku := strings.Title(r.SKU)

	if sku == "Trial" {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: DatabricksWorkspaceUsageSchema,
		}
	}

	var allPurpose, jobs, jobsLight *decimal.Decimal

	if r.MonthlyAllPurposeComputeDbuHrs != nil {
		allPurpose = decimalPtr(decimal.NewFromInt(*r.MonthlyAllPurposeComputeDbuHrs))
	}
	costComponents = append(costComponents, r.databricksCostComponent(
		"All-purpose compute DBUs",
		fmt.Sprintf("%s All-purpose Compute", sku),
		allPurpose,
	))

	if r.MonthlyJobsComputeDbuHrs != nil {
		jobs = decimalPtr(decimal.NewFromInt(*r.MonthlyJobsComputeDbuHrs))
	}
	costComponents = append(costComponents, r.databricksCostComponent(
		"Jobs compute DBUs",
		fmt.Sprintf("%s Jobs Compute", sku),
		jobs,
	))

	if r.MonthlyJobsLightComputeDbuHrs != nil {
		jobsLight = decimalPtr(decimal.NewFromInt(*r.MonthlyJobsLightComputeDbuHrs))
	}
	costComponents = append(costComponents, r.databricksCostComponent(
		"Jobs light compute DBUs",
		fmt.Sprintf("%s Jobs Light Compute", sku),
		jobsLight,
	))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: DatabricksWorkspaceUsageSchema,
	}
}

func (r *DatabricksWorkspace) databricksCostComponent(name, skuName string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "DBU-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
