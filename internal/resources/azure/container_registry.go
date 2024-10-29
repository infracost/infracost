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

func (r *ContainerRegistry) CoreType() string {
	return "ContainerRegistry"
}

func (r *ContainerRegistry) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_build_vcpu_hrs", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *ContainerRegistry) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ContainerRegistry) BuildResource() *schema.Resource {

	var locationsCount int
	var storageGB, monthlyBuildVCPU *decimal.Decimal
	var overStorage decimal.Decimal

	sku := "Classic"
	includedStorage := decimal.NewFromFloat(10)

	if r.SKU != "" {
		sku = r.SKU
	}

	switch sku {
	case "Basic":
		includedStorage = decimal.NewFromFloat(10)
	case "Standard":
		includedStorage = decimal.NewFromFloat(100)
	case "Premium":
		includedStorage = decimal.NewFromFloat(500)
	}

	locationsCount = r.GeoReplicationLocations

	costComponents := make([]*schema.CostComponent, 0)

	if locationsCount > 0 {
		suffix := fmt.Sprintf("%d locations", locationsCount)
		if locationsCount == 1 {
			suffix = fmt.Sprintf("%d location", locationsCount)
		}
		costComponents = append(costComponents, r.containerRegistryGeolocationCostComponent(fmt.Sprintf("Geo replication (%s)", suffix), sku))
	}

	costComponents = append(costComponents, r.containerRegistryCostComponent(fmt.Sprintf("Registry usage (%s)", sku), sku))

	if r.StorageGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
		if storageGB.GreaterThan(includedStorage) {
			overStorage = storageGB.Sub(includedStorage)
			storageGB = &overStorage
			costComponents = append(costComponents, r.containerRegistryStorageCostComponent(fmt.Sprintf("Storage (over %sGB)", includedStorage), sku, storageGB))
		}
	} else {
		costComponents = append(costComponents, r.containerRegistryStorageCostComponent(fmt.Sprintf("Storage (over %sGB)", includedStorage), sku, storageGB))
	}

	if r.MonthlyBuildVCPUHrs != nil {
		monthlyBuildVCPU = decimalPtr(decimal.NewFromFloat(*r.MonthlyBuildVCPUHrs * 3600))
	}

	costComponents = append(costComponents, r.containerRegistryCPUCostComponent("Build vCPU", sku, monthlyBuildVCPU))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ContainerRegistry) containerRegistryCostComponent(name, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "days",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(30)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
func (r *ContainerRegistry) containerRegistryGeolocationCostComponent(name, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "days",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(30 * int64(r.GeoReplicationLocations))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
func (r *ContainerRegistry) containerRegistryStorageCostComponent(name, sku string, storage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storage,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
		UsageBased: true,
	}
}
func (r *ContainerRegistry) containerRegistryCPUCostComponent(name, sku string, monthlyBuildVCPU *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            "seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyBuildVCPU,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
		UsageBased: true,
	}
}
