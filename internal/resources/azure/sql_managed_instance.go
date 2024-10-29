package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

const (
	sqlMIServiceName   = "SQL Managed Instance"
	sqlMIProductFamily = "Databases"
)

// *** this resource is deprecated in v3.0 of AzureRM provider and will be removed in v4.0 ***
// SQLManagedInstance struct represents an azure Sql Managed Instance.
//
// # SQLManagedInstance currently only Gen5 database instance
//
// More resource information here: https://azure.microsoft.com/en-gb/products/azure-sql/managed-instance/
// Pricing information here: https://azure.microsoft.com/en-gb/pricing/details/azure-sql-managed-instance/single/
type SQLManagedInstance struct {
	Address            string
	Region             string
	SKU                string
	LicenseType        string
	Cores              int64
	StorageSizeInGb    int64
	StorageAccountType string
	// LongTermRetentionStorageGB defines a usage param that allows users to define how many gb of cold storage the database uses.
	// This is storage that can be kept for up to 10 years.
	LongTermRetentionStorageGB *int64 `infracost_usage:"long_term_retention_storage_gb"`
	BackupStorageGB            *int64 `infracost_usage:"backup_storage_gb"`
}

func (r *SQLManagedInstance) CoreType() string {
	return "SQLManagedInstance"
}

func (r *SQLManagedInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "long_term_retention_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the SQLManagedInstance.
// It uses the `infracost_usage` struct tags to populate data into the SQLManagedInstance.
func (r *SQLManagedInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SQLManagedInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SQLManagedInstance) BuildResource() *schema.Resource {
	costComponents := r.costComponents()

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *SQLManagedInstance) costComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Compute (%s %d Cores)", strings.ToTitle(r.SKU), r.Cores),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr(sqlMIServiceName),
				ProductFamily: strPtr(sqlMIProductFamily),
				AttributeFilters: ([]*schema.AttributeFilter{
					{Key: "productName", Value: r.productDescription()},
					{Key: "skuName", Value: r.meteredName()},
				}),
			},
			PriceFilter: priceFilterConsumption,
		},
	}

	costComponents = append(costComponents, r.sqlMIStorageCostComponent(), r.sqlMIBackupCostComponent())

	if r.LicenseType == "LicenseIncluded" {
		costComponents = append(costComponents, r.sqlMILicenseCostComponent())
	}

	costComponents = append(costComponents, r.sqlMILongTermRetentionStorageGBCostComponent())

	return costComponents
}

func (r *SQLManagedInstance) productDescription() *string {
	productDescription := ""

	if strings.Contains(r.SKU, "GP") {
		productDescription = "SQL Managed Instance General Purpose"
	} else if strings.Contains(r.SKU, "BC") {
		productDescription = "SQL Managed Instance Business Critical"
	}

	if strings.Contains(r.SKU, "Gen5") {
		productDescription = fmt.Sprintf("%s - %s", productDescription, "Compute Gen5")
	}

	return strPtr(productDescription)
}

func (r *SQLManagedInstance) meteredName() *string {
	meterName := fmt.Sprintf("%d %s", r.Cores, "vCore")

	return strPtr(meterName)
}

func (r *SQLManagedInstance) sqlMIStorageCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(r.StorageSizeInGb - 32)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: ([]*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance General Purpose - Storage")},
				{Key: "skuName", Value: strPtr("General Purpose")},
				{Key: "meterName", ValueRegex: regexPtr("Data Stored$")},
			}),
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLManagedInstance) sqlMIBackupCostComponent() *schema.CostComponent {
	var backup *decimal.Decimal

	if r.BackupStorageGB != nil {
		backup = decimalPtr(decimal.NewFromInt(*r.BackupStorageGB))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("PITR backup storage (%s)", r.StorageAccountType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backup,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: ([]*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance PITR Backup Storage")},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Data Stored", r.StorageAccountType))},
			}),
		},
		PriceFilter: priceFilterConsumption,
		UsageBased:  true,
	}
}

func (r *SQLManagedInstance) sqlMILicenseCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "SQL license",
		Unit:           "vCore-hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.Cores)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr("Global"),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance General Purpose - SQL License")},
				{Key: "meterName", Value: strPtr("vCore")},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLManagedInstance) sqlMILongTermRetentionStorageGBCostComponent() *schema.CostComponent {
	var retention *decimal.Decimal

	if r.LongTermRetentionStorageGB != nil {
		retention = decimalPtr(decimal.NewFromInt(*r.LongTermRetentionStorageGB))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("LTR backup storage (%s)", r.StorageAccountType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: retention,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance - LTR Backup Storage")},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("LTR Backup %s Data Stored", r.StorageAccountType))},
			},
		},
		PriceFilter: priceFilterConsumption,
		UsageBased:  true,
	}
}
