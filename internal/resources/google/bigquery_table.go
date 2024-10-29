package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type BigQueryTable struct {
	Address                   string
	Region                    string
	MonthlyStreamingInsertsMB *float64 `infracost_usage:"monthly_streaming_inserts_mb"`
	MonthlyStorageWriteAPIGB  *float64 `infracost_usage:"monthly_storage_write_api_gb"`
	MonthlyStorageReadAPITB   *float64 `infracost_usage:"monthly_storage_read_api_tb"`
	MonthlyActiveStorageGB    *float64 `infracost_usage:"monthly_active_storage_gb"`
	MonthlyLongTermStorageGB  *float64 `infracost_usage:"monthly_long_term_storage_gb"`
}

func (r *BigQueryTable) CoreType() string {
	return "BigQueryTable"
}

func (r *BigQueryTable) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_streaming_inserts_mb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_storage_write_api_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_storage_read_api_tb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_active_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_long_term_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *BigQueryTable) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *BigQueryTable) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.activeStorageCostComponent(),
		r.longTermStorageCostComponent(),
		r.streamingInsertsCostComponent(),
	}

	storageWriteAPICostComponent := r.storageWriteAPICostComponent()
	if storageWriteAPICostComponent != nil {
		costComponents = append(costComponents, storageWriteAPICostComponent)
	}

	storageReadAPICostComponent := r.storageReadAPICostComponent()
	if storageReadAPICostComponent != nil {
		costComponents = append(costComponents, storageReadAPICostComponent)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *BigQueryTable) activeStorageCostComponent() *schema.CostComponent {
	var activeStorageGB *decimal.Decimal
	if r.MonthlyActiveStorageGB != nil {
		activeStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyActiveStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Active storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: activeStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("BigQuery"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(fmt.Sprintf("Active Logical Storage (%s)", r.Region))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("10"),
		},
		UsageBased: true,
	}
}

func (r *BigQueryTable) longTermStorageCostComponent() *schema.CostComponent {
	var longTermStorageGB *decimal.Decimal
	if r.MonthlyLongTermStorageGB != nil {
		longTermStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyLongTermStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Long-term storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: longTermStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("BigQuery"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(fmt.Sprintf("Long Term Logical Storage (%s)", r.Region))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("10"),
		},
		UsageBased: true,
	}
}

func (r *BigQueryTable) streamingInsertsCostComponent() *schema.CostComponent {
	var streamingInsertsMB *decimal.Decimal
	if r.MonthlyStreamingInsertsMB != nil {
		streamingInsertsMB = decimalPtr(decimal.NewFromFloat(*r.MonthlyStreamingInsertsMB))
	}

	return &schema.CostComponent{
		Name:            "Streaming inserts",
		Unit:            "MB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: streamingInsertsMB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("BigQuery"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(fmt.Sprintf("Streaming Insert (%s)", r.Region))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *BigQueryTable) storageWriteAPICostComponent() *schema.CostComponent {
	var storageWriteAPIGB *decimal.Decimal
	if r.MonthlyStorageWriteAPIGB != nil {
		storageWriteAPIGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageWriteAPIGB))
	}

	region := r.mapRegion()
	if region == "" {
		return nil
	}

	return &schema.CostComponent{
		Name:            "Storage write API",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageWriteAPIGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("BigQuery Storage API"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(fmt.Sprintf("BigQuery Storage API - Write (%s)", region))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("2048"),
		},
		UsageBased: true,
	}
}

func (r *BigQueryTable) storageReadAPICostComponent() *schema.CostComponent {
	var storageReadPITB *decimal.Decimal
	if r.MonthlyStorageReadAPITB != nil {
		storageReadPITB = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageReadAPITB))
	}

	region := r.mapRegion()
	if region == "" {
		return nil
	}

	return &schema.CostComponent{
		Name:            "Storage read API",
		Unit:            "TB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageReadPITB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("BigQuery Storage API"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("BigQuery Storage API - Read")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *BigQueryTable) mapRegion() string {
	if strings.HasPrefix(strings.ToLower(r.Region), "us") {
		return "us"
	}
	if strings.HasPrefix(strings.ToLower(r.Region), "europe") {
		return "europe"
	}

	return ""
}
