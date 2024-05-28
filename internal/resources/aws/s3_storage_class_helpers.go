package aws

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

func s3StorageCostComponent(name string, service string, region string, usageType string, storageGB *float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(storageGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func s3StorageVolumeTypeCostComponent(name string, service string, region string, usageType string, volumeType string, storageGB *float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(storageGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "volumeType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", volumeType))},
				{Key: "operation", Value: strPtr("")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func s3ApiCostComponent(name string, service string, region string, usageType string, requests *int64) *schema.CostComponent {
	return s3ApiOperationCostComponent(name, service, region, usageType, "", requests)
}

func s3ApiOperationCostComponent(name string, service string, region string, usageType string, operation string, requests *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1k requests",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: intPtrToDecimalPtr(requests),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/%s/i", operation))},
			},
		},
		UsageBased: true,
	}
}

func s3DataCostComponent(name string, service string, region string, usageType string, dataGB *float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(dataGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func s3DataGroupCostComponent(name string, service string, region string, usageType string, group string, dataGB *float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(dataGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "group", ValueRegex: strPtr(fmt.Sprintf("/%s/i", group))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func s3LifecycleTransitionsCostComponent(region string, usageType string, operation string, requests *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Lifecycle transition",
		Unit:            "1k requests",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: intPtrToDecimalPtr(requests),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", usageType))},
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", operation))},
			},
		},
		UsageBased: true,
	}
}

func s3MonitoringCostComponent(region string, objects *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Monitoring and automation",
		Unit:            "1k objects",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: intPtrToDecimalPtr(objects),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/Monitoring-Automation-INT/")},
			},
		},
		UsageBased: true,
	}
}
