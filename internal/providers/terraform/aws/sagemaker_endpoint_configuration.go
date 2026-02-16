package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func getSagemakerEndpointConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_sagemaker_endpoint_configuration",
		RFunc:               NewSageMakerEndpointConfiguration,
		ReferenceAttributes: []string{},
	}
}

func NewSageMakerEndpointConfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	var costComponents []*schema.CostComponent

	if d.Get("production_variants").Exists() {
		for _, variant := range d.Get("production_variants").Array() {
			costComponents = append(costComponents, sagemakerVariantComponents(region, &variant, u, "Inference instance")...)
		}
	}

	if d.Get("shadow_production_variants").Exists() {
		for _, variant := range d.Get("shadow_production_variants").Array() {
			costComponents = append(costComponents, sagemakerVariantComponents(region, &variant, u, "Shadow instance")...)
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func sagemakerVariantComponents(region string, variant *gjson.Result, u *schema.UsageData, label string) []*schema.CostComponent {
	serverlessConfig := variant.Get("serverless_config")

	if serverlessConfig.Exists() && len(serverlessConfig.Array()) > 0 {
		return sagemakerServerlessComponents(region, serverlessConfig.Array()[0], u)
	}

	return sagemakerInstanceComponents(region, *variant, label)
}

func sagemakerServerlessComponents(region string, config gjson.Result, u *schema.UsageData) []*schema.CostComponent {
	var components []*schema.CostComponent
	memorySizeMB := config.Get("memory_size_in_mb").Int()
	provisionedConcurrencyCount := config.Get("provisioned_concurrency").Int()
	regionPrefix := regionToUsagePrefix(region)

	if u != nil && u.Get("monthly_data_processed_gb").Exists() {
		monthlyData := decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float())
		components = append(components, &schema.CostComponent{
			Name:            "Data processed",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &monthlyData,
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(region),
				Service:    strPtr("AmazonSageMaker"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "operation", Value: strPtr("Invoke-Endpoint")},
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s-Hst:Data-Bytes-Out/", regionPrefix))},
				},
			},
		})
	}

	// COMPUTE DURATION (Standard execution)
	monthlyDuration := decimal.NewFromFloat(u.Get("monthly_inference_duration_seconds").Float())
	components = append(components, &schema.CostComponent{
		Name:            fmt.Sprintf("Compute (%vMB)", memorySizeMB),
		Unit:            "seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: &monthlyDuration,
		UsageBased:      true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonSageMaker"),
			ProductFamily: strPtr("ML Serverless"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ServerlessInf:Mem-%vGB/", memorySizeMB/1024))},
			},
		},
	})

	// 3. PROVISIONED CONCURRENCY (If enabled)
	if provisionedConcurrencyCount > 0 {
		// PC Readiness (Warm slots) - Billed 24/7
		const monthlyHours = 730
		const hourlySeconds = 3600
		pcQuantity := decimal.NewFromInt(provisionedConcurrencyCount * monthlyHours * hourlySeconds)
		components = append(components, &schema.CostComponent{
			Name:            "Provisioned concurrency (warm)",
			Unit:            "seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &pcQuantity,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s-ProvisionedConcurrency:Mem-%vGB/", regionPrefix, memorySizeMB/1024))},
				},
			},
		})

		// PC Execution (Billed when request hits a warm slot)
		pcUsage := decimal.NewFromFloat(u.Get("monthly_provisioned_concurrency_inference_duration_seconds").Float())
		components = append(components, &schema.CostComponent{
			Name:            "Provisioned concurrency execution",
			Unit:            "seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &pcUsage,
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s-ProvisionedConcurrency:Usage-%vGB/", regionPrefix, memorySizeMB/1024))},
				},
			},
		})
	}

	return components
}

func sagemakerInstanceComponents(region string, variant gjson.Result, label string) []*schema.CostComponent {
	instanceType := variant.Get("instance_type").String()
	count := decimal.NewFromInt(variant.Get("initial_instance_count").Int())

	if count.IsZero() {
		count = decimal.NewFromInt(1)
	}

	components := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("%s (%s)", label, instanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: &count,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(instanceType)},
				},
			},
		},
	}

	volumeSize := variant.Get("volume_size_in_gb").Int()
	if volumeSize > 0 {
		monthlyQty := count.Mul(decimal.NewFromInt(volumeSize))
		components = append(components, &schema.CostComponent{
			Name:            fmt.Sprintf("%s storage", label),
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &monthlyQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{
						Key:        "usagetype",
						ValueRegex: strPtr(fmt.Sprintf("/%s-Studio:VolumeUsage.gp3/", regionToUsagePrefix(region))), //USE1-Studio:VolumeUsage.gp3
					},
				},
			},
		})
	}

	return components
}

func regionToUsagePrefix(region string) string {
	mapping := map[string]string{
		"us-east-1":      "USE1",
		"us-east-2":      "USE2",
		"us-west-1":      "USW1",
		"us-west-2":      "USW2",
		"eu-central-1":   "EUC1",
		"eu-west-1":      "EUW1",
		"eu-west-2":      "EUW2",
		"ap-southeast-1": "APS1",
		"ap-southeast-2": "APS2",
		"ap-northeast-1": "APN1",
		"ap-northeast-2": "APN2",
		"ca-central-1":   "CAN1",
		"sa-east-1":      "SAE1",
	}

	if prefix, ok := mapping[region]; ok {
		return prefix
	}
	return "" // Default to empty or handle error
}
