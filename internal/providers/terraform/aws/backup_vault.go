package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getBackupVaultRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_backup_vault",
		CoreRFunc: NewBackupVault,
		Notes:     []string{"AWS Storage Gateway Volume Backup prices could not be found in the AWS pricing data."},
	}
}
func NewBackupVault(d *schema.ResourceData) schema.CoreResource {
	r := &aws.BackupVault{Address: d.Address, Region: d.Get("region").String()}
	return r
}
