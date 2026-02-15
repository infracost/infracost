package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func getSagemakerEndpointConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sagemaker_endpoint_configuration",
		RFunc: NewSageMakerEndpointConfiguration,
		// We leave this empty for now. If we later want to support
		// Inference Pipelines that pull costs from specific models,
		// we would add "production_variants.model_name" here.
		ReferenceAttributes: []string{},
	}
}

func NewSageMakerEndpointConfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	var costComponents []*schema.CostComponent

	// 1. Process standard Production Variants
	if d.Get("production_variants").Exists() {
		for _, variant := range d.Get("production_variants").Array() {
			costComponents = append(costComponents, sagemakerVariantComponents(region, &variant, u, "Inference instance")...)
		}
	}

	// 2. Process Shadow Production Variants (usually use provisioned instances)
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

	// If it's Serverless
	if serverlessConfig.Exists() && len(serverlessConfig.Array()) > 0 {
		return sagemakerServerlessComponents(region, serverlessConfig.Array()[0], u)
	}

	// Otherwise, it's Provisioned (using your existing helper)
	return sagemakerInstanceComponents(region, *variant, label)
}

func sagemakerServerlessComponents(region string, config gjson.Result, u *schema.UsageData) []*schema.CostComponent {
	var components []*schema.CostComponent
	memorySizeMB := config.Get("memory_size_in_mb").Int()
	pcCount := config.Get("provisioned_concurrency").Int()
	regionPrefix := regionToUsagePrefix(region)

	// 1. DATA PROCESSING (IN & OUT)
	// Covers the $0.016/GB fee you discovered
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

	// 2. COMPUTE DURATION (Standard execution)
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
	if pcCount > 0 {
		// PC Readiness (Warm slots) - Billed 24/7 (730 hours/month)
		pcQuantity := decimal.NewFromInt(pcCount * 730 * 3600) // slots * hours * seconds
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
		// Usually separate usagetype: ServerlessInf:PC-Usage
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

// Fix 2: Change parameter type to gjson.Result to match the loop output
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
			// Fix 3: Infracost uses &count (pointer to decimal) instead of decimal.Ptr
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
						Key: "usagetype",
						// We use a wildcard/suffix match for the usage type
						// so it works in us-east-1 (USE1) and others.
						ValueRegex: strPtr("/Studio:VolumeUsage.gp3/"),
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
