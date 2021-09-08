package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeRegionInstanceGroupManagerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_region_instance_group_manager",
		RFunc:               NewComputeRegionInstanceGroupManager,
		Notes:               []string{"Multiple versions are not supported."},
		ReferenceAttributes: []string{"version.0.instance_template"},
	}
}

func NewComputeRegionInstanceGroupManager(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	machineType := getMachineType(d)
	purchaseOption := getPurchaseOption(d)

	targetSize := decimal.NewFromInt(1)
	if d.Get("target_size").Exists() {
		targetSize = decimal.NewFromInt(d.Get("target_size").Int())
	}

	costComponents := []*schema.CostComponent{computeCostComponent(region, machineType, purchaseOption, targetSize)}

	diskCostComponents := getDisksFromTemplate(d, region, targetSize)
	costComponents = append(costComponents, diskCostComponents...)

	guestAcceleratorComponents := getAcceleratorsFromTemplate(d, region, purchaseOption, targetSize)
	costComponents = append(costComponents, guestAcceleratorComponents...)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
