package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

const (
	logAnalyticsServiceName = "Log Analytics"
	azureMonitorServiceName = "Azure Monitor"
	governanceProductFamily = "Management and Governance"

	skuCapacityReservation = "CapacityReservation"
	skuPerGB2018           = "PerGB2018"
	skuFree                = "Free"
	skuFilterPAYG          = "Pay-as-you-go"

	logRetentionFreeTierLimit = 30
)

var (
	// unsupportedLegacySkus represents skus that Infracost doesn't support because these skus are
	// legacy pricing tiers: https://docs.microsoft.com/en-us/azure/azure-monitor//logs/manage-cost-storage#legacy-pricing-tiers
	unsupportedLegacySkus = map[string]struct{}{
		"unlimited": {},
		"standard":  {},
		"premium":   {},
		"pernode":   {},
	}
)

// LogAnalyticsWorkspace struct represents an Azure Monitor log workspace. A workspace consolidates data
// from multiple sources into a single data lake. A workspace defines:
//
//		1. The geographic location of the data.
//		2. Access rights that define which users can access data.
//		3. Configuration settings such as the pricing tier and data retention.
//
// Resource information: https://azure.microsoft.com/en-gb/services/monitor/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/monitor/
type LogAnalyticsWorkspace struct {
	Address string
	Region  string
	SKU     string

	ReservationCapacityInGBPerDay int64
	RetentionInDays               int64

	MonthlyLogDataIngestionGB           *float64 `infracost_usage:"monthly_log_data_ingestion_gb"`
	MonthlyAdditionalLogDataRetentionGB *float64 `infracost_usage:"monthly_additional_log_data_retention_gb"`
	MonthlyLogDataExportGB              *float64 `infracost_usage:"monthly_log_data_export_gb"`
}

// LogAnalyticsWorkspaceUsageSchema defines a list which represents the usage schema of LogAnalyticsWorkspace.
var LogAnalyticsWorkspaceUsageSchema = []*schema.UsageItem{
	{
		Key:          "monthly_log_data_ingestion_gb",
		DefaultValue: 0,
		ValueType:    schema.Float64,
	},
	{
		Key:          "monthly_additional_log_data_retention_gb",
		DefaultValue: 0,
		ValueType:    schema.Float64,
	},
	{
		Key:          "monthly_log_data_export_gb",
		DefaultValue: 0,
		ValueType:    schema.Float64,
	},
}

// PopulateUsage parses the u schema.UsageData into the LogAnalyticsWorkspace.
// It uses the `infracost_usage` struct tags to populate data into the LogAnalyticsWorkspace.
func (r *LogAnalyticsWorkspace) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid LogAnalyticsWorkspace struct.
// The returned schema.Resource can have 3 potential schema.CostComponent associated with it:
//
//		1. Log data ingestion, which can be either:
//			a) Pay-as-you-go, which is only valid for a sku of PerGB2018 and uses a usage param
//			b) Billed per commitment tiers, which is only valid for a sku of CapacityReservation
//		2. Log retention, which is free up to 31 days and then is billed per GB of data retained after the free limit.
//		   Additional GB used comes from the usage param.
//		3. Data export, which is billed per monthly GB exported and is defined from a usage param.
//
// Outside the above rules - if the workspace has sku of Free we return as a free resource & if the workspace sku
// is in a list of unsupported skus then we mark as skipped with a warning.
func (r *LogAnalyticsWorkspace) BuildResource() *schema.Resource {
	if r.SKU == skuFree {
		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			NoPrice:     true,
			UsageSchema: LogAnalyticsWorkspaceUsageSchema,
		}
	}

	if _, ok := unsupportedLegacySkus[strings.ToLower(r.SKU)]; ok {
		log.Warnf("skipping %s as it uses legacy pricing options", r.Address)

		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			UsageSchema: LogAnalyticsWorkspaceUsageSchema,
		}
	}

	var costComponents []*schema.CostComponent

	if r.SKU == skuPerGB2018 && r.MonthlyLogDataIngestionGB != nil {
		costComponents = append(costComponents, r.logDataIngestion())
	}

	if r.SKU == skuCapacityReservation && r.ReservationCapacityInGBPerDay > 0 {
		costComponents = append(costComponents, r.logDataIngestionFromCapacityReservation())
	}

	if r.RetentionInDays > logRetentionFreeTierLimit {
		costComponents = append(costComponents, r.logDataRetention())
	}

	if r.MonthlyLogDataExportGB != nil {
		costComponents = append(costComponents, r.logDataExport())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    LogAnalyticsWorkspaceUsageSchema,
		CostComponents: costComponents,
	}
}

func (r *LogAnalyticsWorkspace) logDataIngestionFromCapacityReservation() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Log data ingestion",
		Unit:            fmt.Sprintf("%d GB (per day)", r.ReservationCapacityInGBPerDay),
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(30)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(azureMonitorServiceName),
			ProductFamily: strPtr(governanceProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(fmt.Sprintf("%d GB Commitment Tier", r.ReservationCapacityInGBPerDay))},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%d GB Commitment Tier", r.ReservationCapacityInGBPerDay))},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *LogAnalyticsWorkspace) logDataIngestion() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Log data ingestion",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(*r.MonthlyLogDataIngestionGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(logAnalyticsServiceName),
			ProductFamily: strPtr(governanceProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(skuFilterPAYG)},
				{Key: "meterName", Value: strPtr("Data Ingestion")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("5"),
		},
	}
}

func (r *LogAnalyticsWorkspace) logDataRetention() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Log data retention",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(*r.MonthlyAdditionalLogDataRetentionGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(logAnalyticsServiceName),
			ProductFamily: strPtr(governanceProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(skuFilterPAYG)},
				{Key: "meterName", Value: strPtr("Data Retention")},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *LogAnalyticsWorkspace) logDataExport() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Log data export",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(*r.MonthlyLogDataExportGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(azureMonitorServiceName),
			ProductFamily: strPtr(governanceProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Log Analytics data export")},
				{Key: "meterName", Value: strPtr("Data Exported")},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}
