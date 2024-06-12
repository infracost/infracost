package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getBackupProtectedVmRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "azurerm_backup_protected_vm",
		ReferenceAttributes: []string{
			"resource_group_name",
			"backup_policy_id",
			"source_vm_id",
			"recovery_vault_name",
		},
		CoreRFunc: func(d *schema.ResourceData) schema.CoreResource {
			return schema.BlankCoreResource{
				Name: d.Address,
				Type: d.Type,
			}
		},
	}
}

// newBackupProtectedVm returns a azure.BackupProtectedVM with attributes parsed from HCL.
// Note: archive tier not supported https://github.com/hashicorp/terraform-provider-azurerm/issues/21051 by the provider.
func newBackupProtectedVm(d *schema.ResourceData) *azure.BackupProtectedVM {
	region := d.Region
	vms := d.References("source_vm_id")
	if len(vms) == 0 {
		logging.Logger.Warn().Msgf("skipping resource %s as cannot find referenced source vm", d.Address)
		return nil
	}

	vm := vms[0]
	var osDiskSizeGB int64 = 128
	if vm.Get("storage_os_disk.0.disk_size_gb").Exists() {
		osDiskSizeGB = vm.Get("storage_os_disk.0.disk_size_gb").Int()
	}

	var dataDiskSizeGB int64 = 0
	for _, dd := range vm.Get("storage_data_disk").Array() {
		dataDiskSizeGB += dd.Get("disk_size_gb").Int()
	}

	diskSizeGB := osDiskSizeGB + dataDiskSizeGB
	storageType := "GRS"
	recoveryVaults := d.References("recovery_vault_name")
	if len(recoveryVaults) > 0 {
		vault := recoveryVaults[0]
		mode := strings.ToLower(vault.Get("storage_mode_type").String())
		switch mode {
		case "locallyredundant":
			storageType = "LRS"
		case "zoneredundant":
			storageType = "ZRS"
		}

		crossRegion := vault.GetBoolOrDefault("cross_region_restore_enabled", false)
		if storageType == "GRS" && crossRegion {
			storageType = "RA-GRS"
		}
	}

	return &azure.BackupProtectedVM{
		Address:     d.Address,
		Region:      region,
		StorageType: storageType,
		DiskSizeGB:  float64(diskSizeGB),
	}
}
