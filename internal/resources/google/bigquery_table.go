package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type BigqueryTable struct {
	Address                   *string
	Region                    *string
	MonthlyStreamingInsertsMb *float64 `infracost_usage:"monthly_streaming_inserts_mb"`
	MonthlyStorageWriteAPIGb  *float64 `infracost_usage:"monthly_storage_write_api_gb"`
	MonthlyStorageReadAPITb   *float64 `infracost_usage:"monthly_storage_read_api_tb"`
	MonthlyActiveStorageGb    *float64 `infracost_usage:"monthly_active_storage_gb"`
	MonthlyLongTermStorageGb  *float64 `infracost_usage:"monthly_long_term_storage_gb"`
}

var BigqueryTableUsageSchema = []*schema.UsageItem{{Key: "monthly_streaming_inserts_mb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_storage_write_api_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_storage_read_api_tb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_active_storage_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_long_term_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *BigqueryTable) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *BigqueryTable) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := []*schema.CostComponent{}

	var activeStorageGB *decimal.Decimal
	if r.MonthlyActiveStorageGb != nil {
		activeStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyActiveStorageGb))
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
	if r.MonthlyLongTermStorageGb != nil {
		longTermStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyLongTermStorageGb))
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
	if r.MonthlyStreamingInsertsMb != nil {
		streamingInsertsMB = decimalPtr(decimal.NewFromFloat(*r.MonthlyStreamingInsertsMb))
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
	if r.MonthlyStorageWriteAPIGb != nil {
		storageWriteAPI = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageWriteAPIGb))
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
	if r.MonthlyStorageReadAPITb != nil {
		storageReadAPI = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageReadAPITb))
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
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: BigqueryTableUsageSchema,
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
