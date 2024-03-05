package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// BackupProtectedVM struct represents a backup for a given VM into a recovery services vault.
//
// Backup information: https://learn.microsoft.com/en-us/azure/backup/backup-overview
// Resource information: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/backup_policy_vm
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/backup/
type BackupProtectedVM struct {
	Address     string
	Region      string
	StorageType string
	DiskSizeGB  float64

	// DiskUtilizationGB is an override that allows users to specify how much
	// data is actually stored on the VM and will be stored in the vault. By
	// default, we assume that the total VM storage capacity will be backed up.
	DiskUtilizationGB *float64 `infracost_usage:"disk_utilization_gb"`
	// AdditionalBackupRetentionGB allows user sto specify how much additional monthly data
	// is stored in the parent vault because of daily/monthly/yearly retention policies.
	// Azure Backup uses incremental backups, which means that after the initial full backup,
	// it only stores the changes made to the data since the last backup.
	//
	// In the future, it might be better to switch this to a percentage which can be used as daily churn of
	// data from the parent vault. We can then infer the data stored using azurerm_backup_policy_vm and the disk
	// utilization. However, attempts were done when initially writing this mapping, and it proved hard to match
	// up to the exact churn & estimated storage that the azure pricing calculator shows.
	AdditionalBackupRetentionGB *float64 `infracost_usage:"additional_backup_retention_gb"`
}

func (r *BackupProtectedVM) CoreType() string {
	return "BackupProtectedVM"
}

func (r *BackupProtectedVM) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "disk_utilization_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "additional_backup_retention_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the BackupProtectedVM.
//
// This method is normally called from the parent RecoveryServicesVault.PopulateUsage method.
func (r *BackupProtectedVM) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid BackupProtectedVM struct.
//
// BackupProtectedVM is charged for the backup data stored for the protected VM:
//
// Firstly, the backup is charged a flat monthly fee for the size of the VM that it is backing up:
//
//	Instance < or = 50 GB 	$5 + storage consumed
//	Instance is > 50 GB but < or = 500 GB 	$10 + storage consumed
//	Instance is > 500 GB 	$10 for each 500 GB increment + storage consumed
//
// Then BackupProtectedVM is charged per GB of data stored in the parent recovery service vault. This
// depends on the amount of data stored within the vault and the type of storage that the vault uses, e.g. LRS vs GRS.
func (r *BackupProtectedVM) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			r.additionalCostForSizeOfVM(),
			r.storageCostsForVM(),
		},
	}
}

func (r *BackupProtectedVM) additionalCostForSizeOfVM() *schema.CostComponent {
	unit := "under 50 GB"
	quantity := decimal.NewFromInt(1)
	filter := &schema.AttributeFilter{
		Key:   "meterName",
		Value: strPtr("Azure Files Protected Instances"),
	}

	utilization := r.diskUtilization()
	if utilization > 50 {
		filter = &schema.AttributeFilter{
			Key:   "meterName",
			Value: strPtr("Azure VM Protected Instances"),
		}
		unit = "under 500 GB"

		if utilization > 500 {
			quantity = decimal.NewFromInt(int64(utilization) / 500)
			unit = "over 500 GB"
		}
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance backup (%s)", unit),
		Unit:            "month",
		UnitMultiplier:  quantity,
		MonthlyQuantity: &quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Backup"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Backup")},
				filter,
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *BackupProtectedVM) storageCostsForVM() *schema.CostComponent {
	quantity := decimal.NewFromInt(r.totalBackupSize())
	dataStored := fmt.Sprintf("%s data stored", r.StorageType)

	return &schema.CostComponent{
		Name:            dataStored,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Backup"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Backup")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s/i", dataStored))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}
}

func (r *BackupProtectedVM) totalBackupSize() int64 {
	return int64(r.diskUtilization() + r.additionalBackupRetention())
}

func (r *BackupProtectedVM) diskUtilization() float64 {
	if r.DiskUtilizationGB != nil {
		return *r.DiskUtilizationGB
	}

	return r.DiskSizeGB
}

func (r *BackupProtectedVM) additionalBackupRetention() float64 {
	if r.AdditionalBackupRetentionGB != nil {
		return *r.AdditionalBackupRetentionGB
	}

	return 0
}
