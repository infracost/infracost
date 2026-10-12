package azure

import (
	"fmt"
	"regexp"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// PostgreSQLFlexibleServer struct represents Azure PostgreSQL Flexible Server resource.
//
// Resource information: https://docs.microsoft.com/en-gb/azure/postgresql/flexible-server/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/postgresql/flexible-server/
type PostgreSQLFlexibleServer struct {
	Address string
	Region  string

	SKU              string
	Tier             string
	InstanceType     string
	InstanceVersion  string
	Storage          int64
	StorageType      string
	StorageIOPS      int64
	StorageMBps      int64
	HighAvailability bool

	AdditionalBackupStorageGB *float64 `infracost_usage:"additional_backup_storage_gb"`
}

// CoreType returns the name of this resource type
func (r *PostgreSQLFlexibleServer) CoreType() string {
	return "PostgreSQLFlexibleServer"
}

// UsageSchema defines a list which represents the usage schema of PostgreSQLFlexibleServer.
func (r *PostgreSQLFlexibleServer) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "additional_backup_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the PostgreSQLFlexibleServer.
// It uses the `infracost_usage` struct tags to populate data into the PostgreSQLFlexibleServer.
func (r *PostgreSQLFlexibleServer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid PostgreSQLFlexibleServer struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *PostgreSQLFlexibleServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.computeCostComponent(),
		r.storageCostComponent(),
	}
	if r.isPremiumV2() {
		costComponents = append(costComponents, r.premiumV2PerformanceCostComponents()...)
	}
	costComponents = append(costComponents, r.backupCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// computeCostComponent returns a cost component for server compute requirements.
func (r *PostgreSQLFlexibleServer) computeCostComponent() *schema.CostComponent {
	attrs := getFlexibleServerFilterAttributes(r.Tier, r.InstanceType, r.InstanceVersion)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Compute (%s)", r.SKU),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(r.nodes()),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/^%s %s (?:-\\s)?%s/i", attrs.ProductName, attrs.TierName, attrs.Series))},
				{Key: "skuName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", attrs.SKUName))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", attrs.MeterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

// storageCostComponent returns a cost component for server's storage.
func (r *PostgreSQLFlexibleServer) storageCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.Storage > 0 {
		// Storage is in MB
		quantity = decimalPtr(decimal.NewFromInt(r.Storage / 1024).Mul(r.nodes()))
	}

	meterName := "Storage Data Stored"
	if r.isPremiumV2() {
		meterName = "Premium SSD v2 Storage Data Stored"
	}

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				// Azure renamed this to "Azure Database for PostgreSQL Flex Server
				// Storage", but the Delos Cloud regions still publish the old name.
				{Key: "productName", ValueRegex: regexPtr("^(Az DB for PostgreSQL Flexible|Azure Database for PostgreSQL Flex) Server Storage$")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
	}
}

// premiumV2PerformanceCostComponents returns cost components for Premium SSD
// v2 IOPS and throughput above the baseline Azure includes for free.
func (r *PostgreSQLFlexibleServer) premiumV2PerformanceCostComponents() []*schema.CostComponent {
	freeIOPS, freeMBps := int64(3000), int64(125)
	if r.Storage/1024 >= 400 {
		freeIOPS, freeMBps = 12000, 500
	}

	var components []*schema.CostComponent
	if r.StorageIOPS > freeIOPS {
		components = append(components, r.premiumV2CostComponent("Provisioned IOPS", "IOPS", "Premium SSD v2 IOPS Provisioned IOPS", r.StorageIOPS-freeIOPS))
	}
	if r.StorageMBps > freeMBps {
		components = append(components, r.premiumV2CostComponent("Provisioned throughput", "MB/s", "Premium SSDv2 Throughput Prov Throughput (MBps)", r.StorageMBps-freeMBps))
	}
	return components
}

func (r *PostgreSQLFlexibleServer) premiumV2CostComponent(name, unit, meterName string, quantity int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(quantity).Mul(r.nodes())),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for PostgreSQL Flex Server Storage")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
	}
}

func (r *PostgreSQLFlexibleServer) isPremiumV2() bool {
	return r.StorageType == "PremiumV2_LRS"
}

// nodes is the number of servers billed: the primary, plus a standby with the
// same compute and storage when high availability is on.
func (r *PostgreSQLFlexibleServer) nodes() decimal.Decimal {
	if r.HighAvailability {
		return decimal.NewFromInt(2)
	}
	return decimal.NewFromInt(1)
}

// backupCostComponent returns a cost component for additional backup storage.
func (r *PostgreSQLFlexibleServer) backupCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.AdditionalBackupStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.AdditionalBackupStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Additional backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for PostgreSQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for PostgreSQL Flexible Server Backup Storage")},
				{Key: "meterName", Value: strPtr("Backup Storage LRS Data Stored")},
			},
		},
		UsageBased: true,
	}
}

// flexibleServerFilterAttributes defines CPAPI filter attributes for compute
// cost component derived from IaC provider's SKU.
type flexibleServerFilterAttributes struct {
	ProductName string
	SKUName     string
	TierName    string
	MeterName   string
	Series      string
}

// getFlexibleServerFilterAttributes returns a struct with CPAPI filter
// attributes based on values extracted from IaC provider's SKU.
func getFlexibleServerFilterAttributes(tier, instanceType, instanceVersion string) flexibleServerFilterAttributes {
	var skuName, meterName, series string

	tierName := map[string]string{
		"b":  "Burstable",
		"gp": "General Purpose",
		"mo": "Memory Optimized",
	}[tier]

	productName := "Azure Database for PostgreSQL Flexible Server"

	if tier == "b" {
		meterName = fmt.Sprintf("%s[ vcore]*", instanceType)
		skuName = instanceType
		series = "BS"
	} else {
		meterName = "vCore"

		coreRegex := regexp.MustCompile(`(\d+)`)
		match := coreRegex.FindStringSubmatch(instanceType)
		cores := match[1]
		skuName = fmt.Sprintf("%s vCore", cores)

		series = coreRegex.ReplaceAllString(instanceType, "") + instanceVersion
	}

	return flexibleServerFilterAttributes{
		ProductName: productName,
		SKUName:     skuName,
		TierName:    tierName,
		MeterName:   meterName,
		Series:      series,
	}
}
