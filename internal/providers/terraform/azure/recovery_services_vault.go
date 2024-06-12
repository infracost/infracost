package azure

import (
	"sort"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getRecoveryServicesVaultRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_recovery_services_vault",
		CoreRFunc: newRecoveryServicesVault,
		ReferenceAttributes: []string{
			"resource_group_name",
			"azurerm_backup_protected_vm.recovery_vault_name",
		},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			name := d.Get("name").String()
			if name != "" {
				return []string{name}
			}

			return nil
		},
	}
}

func newRecoveryServicesVault(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	vms := d.References("azurerm_backup_protected_vm.recovery_vault_name")

	var protectedVMs []*azure.BackupProtectedVM
	for _, vm := range vms {
		protectedVm := newBackupProtectedVm(vm)
		if protectedVm != nil {
			protectedVMs = append(protectedVMs, protectedVm)
		}
	}

	sort.Slice(protectedVMs, func(i, j int) bool {
		return protectedVMs[i].Address < protectedVMs[j].Address
	})

	return &azure.RecoveryServicesVault{
		Address:      d.Address,
		Region:       region,
		ProtectedVMs: protectedVMs,
	}
}
