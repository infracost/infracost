package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_snapshot",
		CoreRFunc: newSnapshot,
		ReferenceAttributes: []string{
			"resource_group_name",
			"source_uri",
		},
	}
}

func newSnapshot(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	return &azure.Image{
		Type:      "Snapshot",
		StorageGB: snapshotStorageSize(d),
		Address:   d.Address,
		Region:    region,
	}
}

func snapshotStorageSize(d *schema.ResourceData) *float64 {
	v := d.Get("disk_size_gb")
	if v.Exists() && v.Value() != nil {
		size := v.Float()
		return &size
	}

	refs := d.References("source_uri")
	if len(refs) > 0 {
		size := refs[0].Get("disk_size_gb").Float()
		return &size
	}

	return nil
}
