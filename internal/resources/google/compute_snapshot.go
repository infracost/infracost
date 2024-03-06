package google

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeSnapshot struct represents Compute Snapshot resource.
type ComputeSnapshot struct {
	Address  string
	Region   string
	DiskSize float64

	// "usage" args
	StorageGB *float64 `infracost_usage:"storage_gb"`
}

func (r *ComputeSnapshot) CoreType() string {
	return "ComputeSnapshot"
}

// UsageSchema defines a list which represents the usage schema of ComputeSnapshot.
func (r *ComputeSnapshot) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the ComputeSnapshot.
// It uses the `infracost_usage` struct tags to populate data into the ComputeSnapshot.
func (r *ComputeSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeSnapshot struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeSnapshot) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.storageCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// storageCostComponent returns a cost component for snapshot storage.
func (r *ComputeSnapshot) storageCostComponent() *schema.CostComponent {
	description := "Storage PD Snapshot"

	size := r.DiskSize
	if r.StorageGB != nil {
		size = *r.StorageGB
	}

	var snapshotDiskSize *decimal.Decimal
	if size > 0 {
		snapshotDiskSize = decimalPtr(decimal.NewFromFloat(size))
	}

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: snapshotDiskSize,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr(description)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("5"),
		},
		UsageBased: true,
	}
}
