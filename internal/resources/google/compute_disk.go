package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeDisk struct represents Compute Disk resource.
type ComputeDisk struct {
	Address string
	Region  string
	Type    string
	Size    float64
}

// ComputeDiskUsageSchema defines a list which represents the usage schema of ComputeDisk.
var ComputeDiskUsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the ComputeDisk.
// It uses the `infracost_usage` struct tags to populate data into the ComputeDisk.
func (r *ComputeDisk) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeDisk struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeDisk) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		computeDiskCostComponent(r.Region, r.Type, r.Size, 1),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ComputeDiskUsageSchema,
		CostComponents: costComponents,
	}
}
