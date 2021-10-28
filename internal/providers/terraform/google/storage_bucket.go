package google

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

func GetStorageBucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_storage_bucket",
		RFunc:               NewStorageBucket,
		ReferenceAttributes: []string{},
	}
}

func NewStorageBucket(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	components := []*schema.CostComponent{
		dataStorage(d, u),
	}
	data := dataRetrieval(d, u)
	if data != nil {
		components = append(components, data)
	}
	components = append(components, operations(d, u)...)
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: components,
		SubResources: []*schema.Resource{
			networkEgress(region, u, "Network egress", "Data transfer", StorageBucketEgress),
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
	switch strings.ToLower(storageClass) {
	case "nearline":
		resourceGroup = "NearlineStorage"
	case "coldline":
		resourceGroup = "ColdlineStorage"
	case "archive":
		resourceGroup = "ArchiveStorage"
	default:
		resourceGroup = "RegionalStorage"
	}
	// Set the resource group to the right value if the location is a multi-region region
	if strings.ToLower(resourceGroup) == "regionalstorage" {
		switch region {
		// Multi-region locations
		case "asia", "eu", "us":
			resourceGroup = "MultiRegionalStorage"
		// Dual-region locations
		case "asia1", "eur4", "nam4":
			// The pricing api treats a dual-region as a multi-region
			resourceGroup = "MultiRegionalStorage"
		}
	}

	// Handling an exceptional naming
	if region == "eu" && strings.ToLower(resourceGroup) == "multiregionalstorage" {
		region = "europe"
	}

	return region, resourceGroup
}

func dataStorage(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	location := d.Get("location").String()
	if location == "" {
		location = "US"
	}

	var quantity *decimal.Decimal
	if u != nil && u.Get("storage_gb").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))
	}
	storageClass := "STANDARD"
	if d.Get("storage_class").Exists() {
		storageClass = d.Get("storage_class").String()
	}

	region, resourceGroup := getDSRegionResourceGroup(location, storageClass)
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Storage (%s)", strings.ToLower(storageClass)),
		Unit:            "GiB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr(region),
			Service:    strPtr("Cloud Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(resourceGroup)},
				{Key: "description", ValueRegex: strPtr("/^(?!.*?\\(Early Delete\\))/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}

func operations(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	var classAQuantity *decimal.Decimal
	if u != nil && u.Get("monthly_class_a_operations").Exists() {
		classAQuantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_class_a_operations").Int()))
	}
	var classBQuantity *decimal.Decimal
	if u != nil && u.Get("monthly_class_b_operations").Exists() {
		classBQuantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_class_b_operations").Int()))
	}
	storageClass := "STANDARD"
	if d.Get("storage_class").Exists() {
		storageClass = d.Get("storage_class").String()
	}

	storageClassResourceGroupMap := map[string]string{
		"STANDARD":       "RegionalOps",
		"REGIONAL":       "RegionalOps",
		"MULTI_REGIONAL": "MultiRegionalOps",
		"NEARLINE":       "NearlineOps",
		"COLDLINE":       "ColdlineOps",
		"ARCHIVE":        "ArchiveOps",
	}

	return []*schema.CostComponent{
		{
			Name:            "Object adds, bucket/object list (class A)",
			Unit:            "10k operations",
			UnitMultiplier:  decimal.NewFromInt(10000),
			MonthlyQuantity: classAQuantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("gcp"),
				Service:    strPtr("Cloud Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr(storageClassResourceGroupMap[storageClass])},
					{Key: "description", ValueRegex: strPtr("/Class A/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr(""),
			},
		},
		{
			Name:            "Object gets, retrieve bucket/object metadata (class B)",
			Unit:            "10k operations",
			UnitMultiplier:  decimal.NewFromInt(10000),
			MonthlyQuantity: classBQuantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("gcp"),
				Service:    strPtr("Cloud Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "resourceGroup", Value: strPtr(storageClassResourceGroupMap[storageClass])},
					{Key: "description", ValueRegex: strPtr("/Class B/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr(""),
			},
		},
	}
}

func dataRetrieval(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_data_retrieval_gb").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_retrieval_gb").Int()))
	}

	storageClass := "STANDARD"
	if d.Get("storage_class").Exists() {
		storageClass = d.Get("storage_class").String()
	}

	storageClassResourceGroupMap := map[string]string{
		"NEARLINE": "NearlineOps",
		"COLDLINE": "ColdlineOps",
		"ARCHIVE":  "ArchiveOps",
	}
	resourceGroup := storageClassResourceGroupMap[storageClass]
	// Skipping standard, regional and multi-regional since they are free
	if resourceGroup == "" {
		return nil
	}

	return &schema.CostComponent{
		Name:            "Data retrieval",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Service:    strPtr("Cloud Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr(resourceGroup)},
				{Key: "description", ValueRegex: strPtr("/Retrieval/")},
			},
		},
	}
}
