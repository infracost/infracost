package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeRegionInstanceGroupManager struct represents Compute Region Instance
// Group Manager resource.
type ComputeRegionInstanceGroupManager struct {
	Address string
	Region  string

	MachineType       string
	PurchaseOption    string
	TargetSize        int64
	Disks             []*ComputeDisk
	GuestAccelerators []*ComputeGuestAccelerator
}

// ComputeRegionInstanceGroupManagerUsageSchema defines a list which represents the usage schema of ComputeRegionInstanceGroupManager.
var ComputeRegionInstanceGroupManagerUsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the ComputeRegionInstanceGroupManager.
// It uses the `infracost_usage` struct tags to populate data into the ComputeRegionInstanceGroupManager.
func (r *ComputeRegionInstanceGroupManager) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeRegionInstanceGroupManager struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeRegionInstanceGroupManager) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		computeCostComponent(r.Region, r.MachineType, r.PurchaseOption, r.TargetSize),
	}

	for _, disk := range r.Disks {
		costComponents = append(costComponents, computeDiskCostComponent(r.Region, disk.Type, disk.Size, r.TargetSize))
	}

	for _, guestAccel := range r.GuestAccelerators {
		costComponents = append(costComponents, guestAcceleratorCostComponent(r.Region, r.PurchaseOption, guestAccel.Type, guestAccel.Count, r.TargetSize))
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ComputeRegionInstanceGroupManagerUsageSchema,
		CostComponents: costComponents,
	}
}
