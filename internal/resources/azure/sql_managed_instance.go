package azure

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const (
	sqlMIServiceName   = "SQL Managed Instance"
	sqlMIProductFamily = "Databases"
)

// SQLManagedInstance struct represents an azure Sql Managed Instance.
//
// The azurerm_sql_managed_instance resource is deprecated in version 3.0 of the AzureRM provider and will be removed in version 4.0.
// Please use the azurerm_mssql_managed_instance resource instead when available in infracost
//
// Only support for Gen5 is available on that resource
//
// More resource information here: https://azure.microsoft.com/en-gb/products/azure-sql/managed-instance/
// Pricing information here: https://azure.microsoft.com/en-gb/pricing/details/azure-sql-managed-instance/single/
type SQLManagedInstance struct {
	Address            string
	Region             string
	SKU                string
	LicenceType        string
	Cores              *int64
	StorageSizeInGb    *int64
	StorageAccountType string
	// LongTermRetentionStorageGB defines a usage param that allows users to define how many gb of cold storage the database uses.
	// This is storage that can be kept for up to 10 years.
	LongTermRetentionStorageGB *int64 `infracost_usage:"long_term_retention_storage_gb"`
	BackupStorageGb            *int64 `infracost_usage:"backup_storage_gb"`
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
		Name: r.Address,
		UsageSchema: []*schema.UsageItem{
			{Key: "backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
			{Key: "long_term_retention_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		},
		CostComponents: costComponents,
	}
}

func (r *SQLManagedInstance) costComponents() []*schema.CostComponent {

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Compute (%s %s Cores)", strings.ToTitle(r.SKU), strconv.FormatInt(*r.Cores, 10)),
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

	if r.BackupStorageGb != nil {
		costComponents = append(costComponents, r.sqlMIStorageCostComponent(), r.sqlMIBackupCostComponent())
	}

	if r.LicenceType == "LicenseIncluded" {
		costComponents = append(costComponents, r.sqlMILicenseCostComponent())
	}
	if r.LongTermRetentionStorageGB != nil {
		costComponents = append(costComponents, r.sqlMILongTermRetentionStorageGBCostComponent())
	}
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
	meterName := fmt.Sprintf("%s %s", strconv.FormatInt(*r.Cores, 10), "vCore")
	return strPtr(meterName)
}

func (r *SQLManagedInstance) sqlMIStorageCostComponent() *schema.CostComponent {

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Storage %s Gb (first 32 Gb include)", strings.ToTitle(strconv.FormatInt(*r.StorageSizeInGb, 10))),
		Unit:            "Unit of 32Gb",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.StorageSizeInGb - 32)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: ([]*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance General Purpose - Storage")},
				{Key: "meterName", Value: strPtr("Data Stored")},
			}),
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLManagedInstance) sqlMIBackupCostComponent() *schema.CostComponent {
	backupCostComponent := schema.CostComponent{}
	backupCostComponent.Name = fmt.Sprintf("Backup Cost for %s Gb (type %s)", strconv.FormatInt(*r.BackupStorageGb, 10), r.StorageAccountType)
	backupCostComponent.Unit = "Gb"
	backupCostComponent.UnitMultiplier = decimal.NewFromInt(1)
	backupCostComponent.MonthlyQuantity = decimalPtr(decimal.NewFromInt(*r.BackupStorageGb))
	backupCostComponent.ProductFilter = &schema.ProductFilter{
		VendorName:    strPtr(vendorName),
		Region:        strPtr(r.Region),
		Service:       strPtr(sqlMIServiceName),
		ProductFamily: strPtr(sqlMIProductFamily),
		AttributeFilters: ([]*schema.AttributeFilter{
			{Key: "productName", Value: strPtr("SQL Managed Instance PITR Backup Storage")},
			{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Data Stored", r.StorageAccountType))},
		}),
	}
	backupCostComponent.PriceFilter = priceFilterConsumption
	return &backupCostComponent
}

func (r *SQLManagedInstance) sqlMILicenseCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "SQL license",
		Unit:           "vCore-hours",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(*r.Cores)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr("Global"),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance General Purpose - SQL License")},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLManagedInstance) sqlMILongTermRetentionStorageGBCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Long Term Retention Storage Backup (%s Gb (type %s))", strconv.FormatInt(*r.LongTermRetentionStorageGB, 10), r.StorageAccountType),
		Unit:            "Gb/Month",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.LongTermRetentionStorageGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr(sqlMIServiceName),
			ProductFamily: strPtr(sqlMIProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("SQL Managed Instance - LTR Backup Storage")},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("Backup %s Data Stored", r.StorageAccountType))},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}
