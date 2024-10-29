package azure

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getImageRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_image",
		CoreRFunc: newImage,
		ReferenceAttributes: []string{
			"resource_group_name",
			"source_virtual_machine_id",
			"os_disk.0.managed_disk_id",
			"data_disk.#.managed_disk_id",
		},
	}
}

func newImage(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	return &azure.Image{
		StorageGB: imageStorageSize(d),
		Address:   d.Address,
		Region:    region,
	}
}

func imageStorageSize(d *schema.ResourceData) *float64 {
	diskSize := getImageDiskStorage(d)

	sources := d.References("source_virtual_machine_id")
	if diskSize == 0 && len(sources) > 0 {
		diskSize += getVMStorageSize(sources[0])
	}

	if diskSize == 0 {
		return nil
	}

	return &diskSize
}

func getVMStorageSize(d *schema.ResourceData) float64 {
	var size float64 = 128
	if d.Get("storage_os_disk.0.disk_size_gb").Exists() {
		size = d.Get("storage_os_disk.0.disk_size_gb").Float()
	}

	for _, dd := range d.Get("storage_data_disk").Array() {
		size += dd.Get("disk_size_gb").Float()
	}

	return size
}

func getImageDiskStorage(d *schema.ResourceData) float64 {
	var diskSize float64
	osDisk := d.Get("os_disk.0")
	if osDisk.Exists() {
		refs := d.References("os_disk.0.managed_disk_id")

		diskSize += getDiskSizeGB(osDisk, refs, 0)
	}

	disks := d.Get("data_disk").Array()
	refs := d.References("data_disk.#.managed_disk_id")

	for i, disk := range disks {
		diskSize += getDiskSizeGB(disk, refs, i)
	}

	return diskSize
}

func getDiskSizeGB(disk gjson.Result, refs []*schema.ResourceData, i int) float64 {
	if disk.Get("size_gb").Exists() && disk.Get("size_gb").Value() != nil {
		return disk.Get("size_gb").Float()
	}

	if disk.Get("managed_disk_id").Exists() && len(refs) > i {
		ref := refs[i]
		return ref.Get("disk_size_gb").Float()
	}

	return 0
}
