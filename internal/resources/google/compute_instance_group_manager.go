package google

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeInstanceGroupManager struct represents Compute Instance Group Manager
// resource.
type ComputeInstanceGroupManager struct {
	Address string
	Region  string

	MachineType       string
	PurchaseOption    string
	TargetSize        int64
	Disks             []*ComputeDisk
	ScratchDisks      int
	GuestAccelerators []*ComputeGuestAccelerator
}

func (r *ComputeInstanceGroupManager) CoreType() string {
	return "ComputeInstanceGroupManager"
}

// UsageSchema defines a list which represents the usage schema of ComputeInstanceGroupManager.
func (r *ComputeInstanceGroupManager) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the ComputeInstanceGroupManager.
// It uses the `infracost_usage` struct tags to populate data into the ComputeInstanceGroupManager.
func (r *ComputeInstanceGroupManager) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeInstanceGroupManager struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeInstanceGroupManager) BuildResource() *schema.Resource {
	costComponents, err := computeCostComponents(r.Region, r.MachineType, r.PurchaseOption, r.TargetSize, nil)
	if err != nil {
		logging.Logger.Warn().Msgf("Skipping resource %s. %s", r.Address, err)
		return nil
	}

	subResources := make([]*schema.Resource, 0)

	for _, disk := range r.Disks {
		subResources = append(subResources, disk.BuildResource())
	}

	if r.ScratchDisks > 0 {
		costComponents = append(costComponents, scratchDiskCostComponent(r.Region, r.PurchaseOption, r.ScratchDisks*int(r.TargetSize)))
	}

	for _, guestAccel := range r.GuestAccelerators {
		if component := guestAcceleratorCostComponent(r.Region, r.PurchaseOption, guestAccel.Type, guestAccel.Count, r.TargetSize, nil); component != nil {
			costComponents = append(costComponents, component)
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}
