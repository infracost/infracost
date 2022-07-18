package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type StorageBucket struct {
	Address                     string
	Region                      string
	Location                    string
	StorageClass                string
	StorageGB                   *float64                         `infracost_usage:"storage_gb"`
	MonthlyClassAOperations     *int64                           `infracost_usage:"monthly_class_a_operations"`
	MonthlyClassBOperations     *int64                           `infracost_usage:"monthly_class_b_operations"`
	MonthlyDataRetrievalGB      *float64                         `infracost_usage:"monthly_data_retrieval_gb"`
	MonthlyEgressDataTransferGB *StorageBucketNetworkEgressUsage `infracost_usage:"monthly_egress_data_transfer_gb"`
}

var StorageBucketUsageSchema = []*schema.UsageItem{
	{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_class_a_operations", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "monthly_class_b_operations", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "monthly_data_retrieval_gb", ValueType: schema.Float64, DefaultValue: 0},
	{
		Key:          "monthly_egress_data_transfer_gb",
		ValueType:    schema.SubResourceUsage,
		DefaultValue: &usage.ResourceUsage{Name: "monthly_egress_data_transfer_gb", Items: StorageBucketNetworkEgressUsageSchema},
	},
}

func (r *StorageBucket) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *StorageBucket) BuildResource() *schema.Resource {
	if r.MonthlyEgressDataTransferGB == nil {
		r.MonthlyEgressDataTransferGB = &StorageBucketNetworkEgressUsage{}
	}
	region := r.Region
	components := []*schema.CostComponent{
		dataStorageCostComponent(r.Location, r.StorageClass, r.StorageGB),
	}
	data := dataRetrievalCostComponent(r)
	if data != nil {
		components = append(components, data)
	}
	components = append(components, operationsCostComponents(r.StorageClass, r.MonthlyClassAOperations, r.MonthlyClassBOperations)...)

	r.MonthlyEgressDataTransferGB.Region = region
	r.MonthlyEgressDataTransferGB.Address = "Network egress"
	r.MonthlyEgressDataTransferGB.PrefixName = "Data transfer"
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		SubResources: []*schema.Resource{
			r.MonthlyEgressDataTransferGB.BuildResource(),
		}, UsageSchema: StorageBucketUsageSchema,
	}
}

func getDSRegionResourceGroup(location, storageClass string) (string, string) {

	region := strings.ToLower(location)

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

	if strings.ToLower(resourceGroup) == "regionalstorage" {
		switch region {

		case "asia", "eu", "us":
			resourceGroup = "MultiRegionalStorage"

		case "asia1", "eur4", "nam4":

			resourceGroup = "MultiRegionalStorage"
		}
	}

	if region == "eu" && strings.ToLower(resourceGroup) == "multiregionalstorage" {
		region = "europe"
	}

	return region, resourceGroup
}

func dataStorageCostComponent(location, storageClass string, storageGB *float64) *schema.CostComponent {
	if location == "" {
		location = "US"
	}

	var quantity *decimal.Decimal
	if storageGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*storageGB))
	}
	if storageClass == "" {
		storageClass = "STANDARD"
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
			EndUsageAmount: strPtr(""),
		},
	}
}

func operationsCostComponents(storageClass string, monthlyClassAOperations, monthlyClassBOperations *int64) []*schema.CostComponent {
	var classAQuantity *decimal.Decimal
	if monthlyClassAOperations != nil {
		classAQuantity = decimalPtr(decimal.NewFromInt(*monthlyClassAOperations))
	}
	var classBQuantity *decimal.Decimal
	if monthlyClassBOperations != nil {
		classBQuantity = decimalPtr(decimal.NewFromInt(*monthlyClassBOperations))
	}
	if storageClass == "" {
		storageClass = "STANDARD"
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
					{Key: "description", ValueRegex: regexPtr("^(?:(?!Tagging).)* Class A")},
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
					{Key: "description", ValueRegex: regexPtr("^(?:(?!Tagging).)* Class B")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr(""),
			},
		},
	}
}

func dataRetrievalCostComponent(r *StorageBucket) *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyDataRetrievalGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataRetrievalGB))
	}

	storageClass := "STANDARD"
	if r.StorageClass != "" {
		storageClass = r.StorageClass
	}

	storageClassResourceGroupMap := map[string]string{
		"NEARLINE": "NearlineOps",
		"COLDLINE": "ColdlineOps",
		"ARCHIVE":  "ArchiveOps",
	}
	resourceGroup := storageClassResourceGroupMap[storageClass]

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
