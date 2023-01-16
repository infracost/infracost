package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type BigQueryTable struct {
	Address                         string
	Region                          string
	MonthlyStreamingInsertsMB       *float64 `infracost_usage:"monthly_streaming_inserts_mb"`
	MonthlyStorageWriteAPIGB        *float64 `infracost_usage:"monthly_storage_write_api_gb"`
	MonthlyStorageReadAPITB         *float64 `infracost_usage:"monthly_storage_read_api_tb"`
	MonthlyActiveLogicalStorageGB   *float64 `infracost_usage:"monthly_active_logical_storage_gb"`
	MonthlyActivePhysicalStorageGB  *float64 `infracost_usage:"monthly_active_physical_storage_gb"`
	MonthlyLongTermLogicalStorageGB *float64 `infracost_usage:"monthly_long_term_logical_storage_gb"`
}

var BigQueryTableUsageSchema = []*schema.UsageItem{
	{Key: "monthly_streaming_inserts_mb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_storage_write_api_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_storage_read_api_tb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_active_logical_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_active_physical_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_long_term_logical_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *BigQueryTable) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *BigQueryTable) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.activeLogicalStorageCostComponent(),
		r.activePhysicalStorageCostComponent(),
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
		UsageSchema:    BigQueryTableUsageSchema,
	}
}

func (r *BigQueryTable) activeLogicalStorageCostComponent() *schema.CostComponent {
	var activeStorageGB *decimal.Decimal
	if r.MonthlyActiveLogicalStorageGB != nil {
		activeStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyActiveLogicalStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Active logical storage",
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
	}
}

func (r *BigQueryTable) activePhysicalStorageCostComponent() *schema.CostComponent {
	var activeStorageGB *decimal.Decimal
	if r.MonthlyActiveLogicalStorageGB != nil {
		activeStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyActiveLogicalStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Active physical storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: activeStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("BigQuery"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr(fmt.Sprintf("Active Physical Storage (%s)", r.Region))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("10"),
		},
	}
}

func (r *BigQueryTable) longTermStorageCostComponent() *schema.CostComponent {
	var longTermStorageGB *decimal.Decimal
	if r.MonthlyLongTermLogicalStorageGB != nil {
		longTermStorageGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyLongTermLogicalStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Long-term logical storage",
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
