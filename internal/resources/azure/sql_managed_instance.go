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

// SQLManagedInstance struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://azure.microsoft.com/<PATH/TO/RESOURCE>/
// Pricing information: https://azure.microsoft.com/<PATH/TO/PRICING>/
type SQLManagedInstance struct {
	Address            string
	Region             string
	SKU                string
	LicenceType        string
	Cores              *int64
	StorageSizeInGb    *int64
	StorageAccountType string
	AverageBackupSize  *int64 `infracost_usage:"average_backup_size_gb"`
	WeeklyBackup       *int64 `infracost_usage:"weekly_backup"`
	MonthlyBackup      *int64 `infracost_usage:"monthly_backup"`
	YearlyBackup       *int64 `infracost_usage:"yearly_backup"`
	BackupStorageGb    *int64 `infracost_usage:"backup_storage_gb"`
}

// SQLManagedInstanceUsageSchema defines a list which represents the usage schema of SQLManagedInstance.

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
			{Key: "average_backup_size_gb", DefaultValue: 0, ValueType: schema.Int64},
			{Key: "weekly_backup", DefaultValue: 0, ValueType: schema.Int64},
			{Key: "monthly_backup", DefaultValue: 0, ValueType: schema.Int64},
			{Key: "yearly_backup", DefaultValue: 0, ValueType: schema.Int64},
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
		costComponents = append(costComponents, r.storageCost(), r.backupCost())
	}

	if r.LicenceType == "LicenseIncluded" {
		costComponents = append(costComponents, r.sqlMILicenseCostComponent())
	}
	if r.AverageBackupSize != nil {
		if r.WeeklyBackup != nil && *r.WeeklyBackup > 0 {
			costComponents = append(costComponents, r.sqlMIWeeklyBackupCostComponent())
		}

		if r.MonthlyBackup != nil && *r.MonthlyBackup > 0 {
			costComponents = append(costComponents, r.sqlMIMonthlyBackupCostComponent())
		}

		if r.YearlyBackup != nil && *r.YearlyBackup > 0 {
			costComponents = append(costComponents, r.sqlMIYearlyBackupCostComponent())
		}
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

func (r *SQLManagedInstance) storageCost() *schema.CostComponent {

	storageComponent := schema.CostComponent{}
	storageComponent.Name = fmt.Sprintf("Storage %s Gb (first 32 Gb include)", strings.ToTitle(strconv.FormatInt(*r.StorageSizeInGb, 10)))
	storageComponent.Unit = "Unit of 32Gb"
	storageComponent.UnitMultiplier = decimal.NewFromInt(1)
	storageComponent.MonthlyQuantity = decimalPtr(decimal.NewFromInt(*r.StorageSizeInGb - 32))
	storageComponent.ProductFilter = &schema.ProductFilter{
		VendorName:    strPtr(vendorName),
		Region:        strPtr(r.Region),
		Service:       strPtr(sqlMIServiceName),
		ProductFamily: strPtr(sqlMIProductFamily),
		AttributeFilters: ([]*schema.AttributeFilter{
			{Key: "productName", Value: strPtr("SQL Managed Instance General Purpose - Storage")},
			{Key: "meterName", Value: strPtr("Data Stored")},
		}),
	}

	storageComponent.PriceFilter = priceFilterConsumption
	return &storageComponent
}

func (r *SQLManagedInstance) backupCost() *schema.CostComponent {
	backupCostComponent := schema.CostComponent{}
	backupCostComponent.Name = fmt.Sprintf("Backup Cost for %s Gb %s", strconv.FormatInt(*r.BackupStorageGb, 10), r.StorageAccountType)
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
		UnitMultiplier: decimal.NewFromInt(1),
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

func (r *SQLManagedInstance) sqlMIWeeklyBackupCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Weekly Backup (%s Backups with %s Gb)", strconv.FormatInt(*r.WeeklyBackup, 10), strconv.FormatInt(*r.AverageBackupSize, 10)),
		Unit:            "Gb/Month",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.AverageBackupSize * *r.WeeklyBackup)),
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

func (r *SQLManagedInstance) sqlMIMonthlyBackupCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Monthly Backup (%s Backups with %s Gb)", strconv.FormatInt(*r.MonthlyBackup, 10), strconv.FormatInt(*r.AverageBackupSize, 10)),
		Unit:            "Gb/Month",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.AverageBackupSize * *r.MonthlyBackup)),
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

func (r *SQLManagedInstance) sqlMIYearlyBackupCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Yearly Backup (%s Backups with %s Gb)", strconv.FormatInt(*r.YearlyBackup, 10), strconv.FormatInt(*r.AverageBackupSize, 10)),
		Unit:            "Gb/Month",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(*r.AverageBackupSize * *r.MonthlyBackup)),
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
