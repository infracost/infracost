package google

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeImage struct represents Compute Image resource.
type ComputeImage struct {
	Address     string
	Region      string
	StorageSize float64

	// "usage" args
	StorageGB *float64 `infracost_usage:"storage_gb"`
}

func (r *ComputeImage) CoreType() string {
	return "ComputeImage"
}

// UsageSchema defines a list which represents the usage schema of ComputeImage.
func (r *ComputeImage) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the ComputeImage.
// It uses the `infracost_usage` struct tags to populate data into the ComputeImage.
func (r *ComputeImage) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeImage struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeImage) BuildResource() *schema.Resource {
	storageSize := r.StorageSize
	if r.StorageGB != nil {
		storageSize = *r.StorageGB
	}

	var size *decimal.Decimal
	if storageSize > 0 {
		size = decimalPtr(decimal.NewFromFloat(storageSize))
	}

	costComponents := []*schema.CostComponent{
		storageImageCostComponent(r.Region, "Storage Image", size),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
