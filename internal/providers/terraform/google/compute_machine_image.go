package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeMachineImageRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_machine_image",
		RFunc: NewComputeMachineImage,
	}
}

func NewComputeMachineImage(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	description := "Storage Machine Image"

	var storageSize *decimal.Decimal
	if u != nil && u.Get("storage_gb").Exists() {
		storageSize = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: storageImage(region, description, storageSize),
	}
}
