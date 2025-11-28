package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MySQLFlexibleServer struct represents Azure MySQL Flexible Server resource.
//
// Resource information: https://docs.microsoft.com/en-gb/azure/mysql/flexible-server/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/mysql/flexible-server/
type MySQLFlexibleServer struct {
	Address string
	Region  string

	SKU             string
	Tier            string
	InstanceType    string
	InstanceVersion string
	Storage         int64
	IOPS            int64

	// "usage" args
	AdditionalBackupStorageGB *float64 `infracost_usage:"additional_backup_storage_gb"`
}

// CoreType returns the name of this resource type
func (r *MySQLFlexibleServer) CoreType() string {
	return "MySQLFlexibleServer"
}

// UsageSchema defines a list which represents the usage schema of MySQLFlexibleServerUsageSchema.
func (r *MySQLFlexibleServer) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "additional_backup_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the MySQLFlexibleServer.
// It uses the `infracost_usage` struct tags to populate data into the MySQLFlexibleServer.
func (r *MySQLFlexibleServer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid MySQLFlexibleServer struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MySQLFlexibleServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.computeCostComponent(),
		r.storageCostComponent(),
	}

	if iopsCostComponent := r.iopsCostComponent(); iopsCostComponent != nil {
		costComponents = append(costComponents, iopsCostComponent)
	}

	costComponents = append(costComponents, r.backupCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// computeCostComponent returns a cost component for server compute requirements.
func (r *MySQLFlexibleServer) computeCostComponent() *schema.CostComponent {
	attrs := getFlexibleServerFilterAttributes(r.Tier, r.InstanceType, r.InstanceVersion)

	// MySQL Flexible Server uses "Business Critical" for the Memory Optimized instances
	tierName := attrs.TierName
	if tierName == "Memory Optimized" {
		tierName = "Business Critical"
	}

	series := attrs.Series
	// We've seen two spaces in the data in the past hence '\s+'
	seriesSuffix := fmt.Sprintf("\\s+%s Series", series)
	// This seems to be a special case where the series doesn't appear in the product name
	if tierName == "Business Critical" && series == "Edsv4" {
		seriesSuffix = " Compute"
	}

	if tierName == "General Purpose" && series == "Dadsv5" {
		seriesSuffix = ""
	}

	productNameRegex := fmt.Sprintf("^Azure Database for MySQL Flexible Server %s%s", tierName, seriesSuffix)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Compute (%s)", r.SKU),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for MySQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(productNameRegex)},
				{Key: "skuName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", attrs.SKUName))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", attrs.MeterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

// storageCostComponent returns a cost component for server's storage. If
// storage is not defined, it is assumed it is a minimum default of 20GB.
func (r *MySQLFlexibleServer) storageCostComponent() *schema.CostComponent {
	storage := r.Storage
	if storage == 0 {
		storage = 20 // minimum default
	}

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(storage)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for MySQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for MySQL Flexible Server Storage")},
				{Key: "meterName", Value: strPtr("Storage Data Stored")},
			},
		},
	}
}

// iopsCostComponent returns a cost component for additional IOPS. Each server
// includes free 300 IOPS and 3 IOPS per each storage GB. As minimum storage is
// 20GB, the total free IOPS is 360. If no IOPS is defined it's assumed it is
// the minimum of 360.
func (r *MySQLFlexibleServer) iopsCostComponent() *schema.CostComponent {
	var freeIOPS int64 = 360

	iops := r.IOPS
	if iops == 0 {
		iops = freeIOPS
	}

	additionalIOPS := iops - freeIOPS

	if additionalIOPS <= 0 {
		return nil
	}

	return &schema.CostComponent{
		Name:            "Additional IOPS",
		Unit:            "IOPS",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(additionalIOPS)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Database for MySQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for MySQL Flexible Server Storage")},
				{Key: "skuName", Value: strPtr("Additional IOPS")},
			},
		},
	}
}

// backupCostComponent returns a cost component for additional backup storage.
func (r *MySQLFlexibleServer) backupCostComponent() *schema.CostComponent {
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
			Service:       strPtr("Azure Database for MySQL"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Database for MySQL Flexible Server Backup Storage")},
				{Key: "meterName", Value: strPtr("Backup Storage LRS Data Stored")},
			},
		},
		UsageBased: true,
	}
}
