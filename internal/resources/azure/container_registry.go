package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type ContainerRegistry struct {
	Address                 string
	GeoReplicationLocations int
	Region                  string
	SKU                     string
	StorageGB               *float64 `infracost_usage:"storage_gb"`
	MonthlyBuildVCPUHrs     *float64 `infracost_usage:"monthly_build_vcpu_hrs"`
}

var ContainerRegistryUsageSchema = []*schema.UsageItem{{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_build_vcpu_hrs", ValueType: schema.Float64, DefaultValue: 0}}

func (r *ContainerRegistry) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ContainerRegistry) BuildResource() *schema.Resource {
	region := r.Region

	var locationsCount int
	var storageGb, includedStorage, monthlyBuildVCPU *decimal.Decimal
	var overStorage decimal.Decimal

	sku := "Classic"

	if r.SKU != "" {
		sku = r.SKU
	}

	switch sku {
	case "Classic":
		includedStorage = decimalPtr(decimal.NewFromFloat(10))
	case "Basic":
		includedStorage = decimalPtr(decimal.NewFromFloat(10))
	case "Standard":
		includedStorage = decimalPtr(decimal.NewFromFloat(100))
	case "Premium":
		includedStorage = decimalPtr(decimal.NewFromFloat(500))
	}

	locationsCount = r.GeoReplicationLocations

	costComponents := make([]*schema.CostComponent, 0)

	if locationsCount > 0 {
		suffix := fmt.Sprintf("%d locations", locationsCount)
		if locationsCount == 1 {
			suffix = fmt.Sprintf("%d location", locationsCount)
		}
		costComponents = append(costComponents, ContainerRegistryGeolocationCostComponent(fmt.Sprintf("Geo replication (%s)", suffix), region, sku, locationsCount))
	}

	costComponents = append(costComponents, ContainerRegistryCostComponent(fmt.Sprintf("Registry usage (%s)", sku), region, sku))

	if r.StorageGB != nil {
		storageGb = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
		if storageGb.GreaterThan(*includedStorage) {
			overStorage = storageGb.Sub(*includedStorage)
			storageGb = &overStorage
			costComponents = append(costComponents, ContainerRegistryStorageCostComponent(fmt.Sprintf("Storage (over %sGB)", includedStorage), region, sku, storageGb))
		}
	} else {
		costComponents = append(costComponents, ContainerRegistryStorageCostComponent(fmt.Sprintf("Storage (over %sGB)", includedStorage), region, sku, storageGb))
	}

	if r.MonthlyBuildVCPUHrs != nil {
		monthlyBuildVCPU = decimalPtr(decimal.NewFromFloat(*r.MonthlyBuildVCPUHrs * 3600))
	}

	costComponents = append(costComponents, ContainerRegistryCPUCostComponent("Build vCPU", region, sku, monthlyBuildVCPU))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: ContainerRegistryUsageSchema,
	}
}

func ContainerRegistryCostComponent(name, region, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "days",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(30)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Container Registry"),
			ProductFamily: strPtr("Containers"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Container Registry")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Registry Unit", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func ContainerRegistryGeolocationCostComponent(name, region, sku string, locationsCount int) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "days",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(30 * int64(locationsCount))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Container Registry"),
			ProductFamily: strPtr("Containers"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Container Registry")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Registry Unit", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func ContainerRegistryStorageCostComponent(name, region, sku string, storage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Container Registry"),
			ProductFamily: strPtr("Containers"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Container Registry")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr("Data Stored")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func ContainerRegistryCPUCostComponent(name, region, sku string, monthlyBuildVCPU *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            "seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyBuildVCPU,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Container Registry"),
			ProductFamily: strPtr("Containers"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Container Registry")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr("Task vCPU Duration")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("6000"),
		},
	}
}
