package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetBigqueryTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_bigquery_table",
		RFunc: NewBigqueryTable,
	}
}

func NewBigqueryTable(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := []*schema.CostComponent{}

	var activeStorageGB *decimal.Decimal
	if u != nil && u.Get("monthly_active_storage_gb").Type != gjson.Null {
		activeStorageGB = decimalPtr(decimal.NewFromFloat(u.Get("monthly_active_storage_gb").Float()))
	}
	costComponents = append(costComponents, bigQueryTableCostComponent(
		"Active storage",
		"GB",
		region,
		"BigQuery",
		fmt.Sprintf("Active Storage (%s)", region),
		"10",
		activeStorageGB,
	))

	var longTermStorageGB *decimal.Decimal
	if u != nil && u.Get("monthly_long_term_storage_gb").Type != gjson.Null {
		longTermStorageGB = decimalPtr(decimal.NewFromFloat(u.Get("monthly_long_term_storage_gb").Float()))
	}
	costComponents = append(costComponents, bigQueryTableCostComponent(
		"Long-term storage",
		"GB",
		region,
		"BigQuery",
		fmt.Sprintf("Long Term Storage (%s)", region),
		"10",
		longTermStorageGB,
	))

	var streamingInsertsMB *decimal.Decimal
	if u != nil && u.Get("monthly_streaming_inserts_mb").Type != gjson.Null {
		streamingInsertsMB = decimalPtr(decimal.NewFromFloat(u.Get("monthly_streaming_inserts_mb").Float()))
	}
	costComponents = append(costComponents, bigQueryTableCostComponent(
		"Streaming inserts",
		"MB",
		region,
		"BigQuery",
		fmt.Sprintf("Streaming Insert (%s)", region),
		"0",
		streamingInsertsMB,
	))

	var storageWriteAPI *decimal.Decimal
	if u != nil && u.Get("monthly_storage_write_api_gb").Type != gjson.Null {
		storageWriteAPI = decimalPtr(decimal.NewFromFloat(u.Get("monthly_storage_write_api_gb").Float()))
	}
	if reg := regionMapping(region); reg != "" {
		costComponents = append(costComponents, bigQueryTableCostComponent(
			"Storage write API",
			"GB",
			reg,
			"BigQuery Storage API",
			fmt.Sprintf("BigQuery Storage API - Write (%s)", reg),
			"2048",
			storageWriteAPI,
		))
	}

	var storageReadAPI *decimal.Decimal
	if u != nil && u.Get("monthly_storage_read_api_tb").Type != gjson.Null {
		storageReadAPI = decimalPtr(decimal.NewFromFloat(u.Get("monthly_storage_read_api_tb").Float()))
	}
	if reg := regionMapping(region); reg != "" {
		costComponents = append(costComponents, bigQueryTableCostComponent(
			"Storage read API",
			"TB",
			reg,
			"BigQuery Storage API",
			"BigQuery Storage API - Read",
			"0",
			storageReadAPI,
		))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func bigQueryTableCostComponent(name, unit, region, service, description, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr(service),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(description)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
	}
}

func regionMapping(region string) string {
	if strings.HasPrefix(strings.ToLower(region), "us") {
		return "us"
	}
	if strings.HasPrefix(strings.ToLower(region), "europe") {
		return "europe"
	}

	return ""
}
