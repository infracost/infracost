package google

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeMachineImage struct represents Compute Machine Image resource.
type ComputeMachineImage struct {
	Address string
	Region  string

	// "usage" args
	StorageGB *float64 `infracost_usage:"storage_gb"`
}

func (r *ComputeMachineImage) CoreType() string {
	return "ComputeMachineImage"
}

// UsageSchema defines a list which represents the usage schema of ComputeMachineImage.
func (r *ComputeMachineImage) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the ComputeMachineImage.
// It uses the `infracost_usage` struct tags to populate data into the ComputeMachineImage.
func (r *ComputeMachineImage) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeMachineImage struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeMachineImage) BuildResource() *schema.Resource {
	var storageSize *decimal.Decimal
	if r.StorageGB != nil {
		storageSize = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	costComponents := []*schema.CostComponent{
		storageImageCostComponent(r.Region, "Storage Machine Image", storageSize),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
