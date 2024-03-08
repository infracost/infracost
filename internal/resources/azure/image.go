package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// Image represents a custom image or a snapshot in Azure Cloud.
//
// A custom image in Azure is created from a VHD or a managed disk, and it's used as the base
// for creating new virtual machines. Image, when representing a custom image,
// takes into account the size of the managed disk that backs the image for cost calculation.
//
// A snapshot in Azure is a read-only copy of a disk, created for backup or to create
// new disks with the same data. When representing a snapshot, Image considers
// the used size of the disk snapshot for cost estimation.
//
// The cost of both custom images and snapshots is primarily based on the storage they consume.
//
// Resource information:
// Custom Images: https://docs.microsoft.com/en-us/azure/virtual-machines/windows/capture-image-resource
// Snapshots: https://docs.microsoft.com/en-us/azure/virtual-machines/disks-snapshot
//
// Pricing information:
// Managed Disks: https://azure.microsoft.com/en-us/pricing/details/managed-disks/
// Storage: https://azure.microsoft.com/en-us/pricing/details/storage/
type Image struct {
	Type string

	Address   string
	Region    string
	StorageGB *float64 `infracost_usage:"storage_gb"`
}

// CoreType returns the name of this resource type.
// If no type is specified, it defaults to "Image".
func (r *Image) CoreType() string {
	if r.Type == "" {
		return "Image"
	}

	return r.Type
}

// UsageSchema defines a list which represents the usage schema of Image.
// Currently, it includes a single key "storage_gb" for the storage size in GB.
func (r *Image) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the Image.
// It uses the `infracost_usage` struct tags to populate data into the Image.
func (r *Image) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid Image struct.
// This method is called after the resource is initialised by an IaC provider.
//
// As of writing, snapshot and image costs are all billed under the "Standard HDD Managed Disks" product, even if the
// disks they are backing up are premium or SSD.
func (r *Image) BuildResource() *schema.Resource {
	var size *decimal.Decimal
	if r.StorageGB != nil {
		size = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: size,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Storage"),
					ProductFamily: strPtr("Storage"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "skuName", Value: strPtr("Snapshots LRS")},
						{Key: "meterName", ValueRegex: regexPtr("LRS Snapshots$")},
						{Key: "productName", Value: strPtr("Standard HDD Managed Disks")},
					},
				},
			},
		},
	}
}
