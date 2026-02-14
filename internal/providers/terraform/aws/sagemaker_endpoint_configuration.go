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

	// Fix 1: We pass the gjson.Result directly to the helper
	if d.Get("production_variants").Exists() {
		for _, variant := range d.Get("production_variants").Array() {
			serverlessConfig := variant.Get("serverless_config")

			if serverlessConfig.Exists() && len(serverlessConfig.Array()) > 0 {
				// --- SERVERLESS PATH ---
				config := serverlessConfig.Array()[0]
				memorySizeMB := config.Get("memory_size_in_mb").Int()

				// AWS bills in seconds per your Postman response ($0.00004/sec)
				// We use UnitMultiplier: 1 because the price is already for the whole 2GB chunk
				monthlyDuration := decimal.NewFromFloat(u.Get("monthly_inference_duration_seconds").Float())
				costComponents = append(costComponents, &schema.CostComponent{
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
							{
								Key:        "usagetype",
								ValueRegex: strPtr("/ServerlessInf:Mem-2GB/"),
							},
						},
					},
					PriceFilter: &schema.PriceFilter{
						PurchaseOption: strPtr("on_demand"),
					},
				})
			} else {
				// --- PROVISIONED PATH ---
				// This helper handles the instance type and storage only.
				costComponents = append(costComponents, sagemakerInstanceComponents(region, variant, "Inference instance")...)
			}
		}
	}

	if d.Get("shadow_production_variants").Exists() {
		for _, variant := range d.Get("shadow_production_variants").Array() {
			costComponents = append(costComponents, sagemakerInstanceComponents(region, variant, "Shadow instance")...)
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
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
