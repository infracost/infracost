package google

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetStorageBucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_storage_bucket",
		RFunc:               NewStorageBucket,
		ReferenceAttributes: []string{},
	}
}

func NewStorageBucket(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			dataStorage(d, u),
		},
	}
}

func getDSRegionResourceGroup(location, storageClass string) (string, string) {
	// Storage buckets have a location field that indicates the location of data. You can
	// take a look at https://cloud.google.com/storage/docs/locations for list of locations.
	// There is a complicated mapping for Standard storage class. If the location is a single
	// region, then the resource group is named "RegionalStorage" and for multi-region regions
	// its "MultiRegionalStorage". For other storage classes they are fixed.

	// Convert the location field in terraform to a valid region name
	region := strings.ToLower(location)

	// Get the right resourceGroup for api query
	var resourceGroup string
	switch storageClass {
	case "NEARLINE":
		resourceGroup = "NearlineStorage"
	case "COLDLINE":
		resourceGroup = "ColdlineStorage"
	case "ARCHIVE":
		resourceGroup = "ArchiveStorage"
	default:
		resourceGroup = "RegionalStorage"
	}
	// Set the resource group to the right value if the location is a multi-region region
	if resourceGroup == "RegionalStorage" {
		switch location {
		// Multi-region locations
		case "ASIA", "EU", "US":
			resourceGroup = "MultiRegionalStorage"
		// Dual-region locations
		case "ASIA1", "EUR4", "NAM4":
			// The pricing api treats a dual-region as  a multi-region
			resourceGroup = "MultiRegionalStorage"
		}
	}

	return region, resourceGroup
}

func dataStorage(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	location := d.Get("location").String()
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_gb_storage").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_storage").Int()))
	}
	storageClass := "STANDARD"
	if d.Get("storage_class").Exists() {
		storageClass = d.Get("storage_class").String()
	}

	region, resourceGroup := getDSRegionResourceGroup(location, storageClass)
	return &schema.CostComponent{
		Name:            "Data storage",
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(resourceGroup)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}
