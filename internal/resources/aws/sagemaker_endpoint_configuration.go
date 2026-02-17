package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// See the pricing information here: https://aws.amazon.com/sagemaker/ai/pricing/.
type SageMakerEndpointConfiguration struct {
	Address  string
	Region   string
	Variants []*SageMakerVariant
}

type SageMakerVariant struct {
	Name                   string
	InstanceType           string
	InitialInstanceCount   int64
	IsServerless           bool
	MemorySizeMB           int64
	ProvisionedConcurrency int64
	MaxConcurrency         int64 // Not billed, but good for completeness
	VolumeSizeInGB         int64
	Label                  string

	// "usage" keys
	MonthlyInferenceDurationSeconds                       *float64 `infracost_usage:"monthly_inference_duration_seconds"`
	MonthlyProvisionedConcurrencyInferenceDurationSeconds *float64 `infracost_usage:"monthly_provisioned_concurrency_inference_duration_seconds"`
	DataProcessedGB                                       *float64 `infracost_usage:"monthly_data_processed_gb"`
}

var SageMakerEndpointConfigurationUsageSchema = []*schema.UsageItem{
	{Key: "monthly_inference_duration_seconds", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_provisioned_concurrency_inference_duration_seconds", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
}

func (s *SageMakerEndpointConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(s, u)
}

func (s *SageMakerEndpointConfiguration) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	for _, variant := range s.Variants {
		if variant.IsServerless {
			costComponents = append(costComponents, s.sagemakerServerlessComponents(variant)...)
		} else {
			costComponents = append(costComponents, s.sagemakerInstanceComponents(variant)...)
		}
	}

	return &schema.Resource{
		Name:           s.Address,
		UsageSchema:    SageMakerEndpointConfigurationUsageSchema,
		CostComponents: costComponents,
	}
}

func (s *SageMakerEndpointConfiguration) sagemakerServerlessComponents(v *SageMakerVariant) []*schema.CostComponent {
	var components []*schema.CostComponent

	// 1. Data processed - Billed per GB of data processed by the endpoint
	monthlyDataProcessedGB := floatPtrToDecimalPtr(v.DataProcessedGB)
	components = append(components, &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyDataProcessedGB,
		UsageBased:      true,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(s.Region),
			Service:    strPtr("AmazonSageMaker"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("Invoke-Endpoint")},
				{Key: "usagetype", ValueRegex: strPtr("/Hst:Data-Bytes-Out/")},
			},
		},
	})

	// 2. Compute duration
	memorySizeMB := v.MemorySizeMB
	monthlyDuration := floatPtrToDecimalPtr(v.MonthlyInferenceDurationSeconds)

	components = append(components, &schema.CostComponent{
		Name:            fmt.Sprintf("Compute (%vMB)", memorySizeMB),
		Unit:            "seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyDuration,
		UsageBased:      true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(s.Region),
			Service:       strPtr("AmazonSageMaker"),
			ProductFamily: strPtr("ML Serverless"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ServerlessInf:Mem-%vGB$/", memorySizeMB/1024))},
			},
		},
	})

	// 3. Provisioned concurrency (if enabled)
	provisionedConcurrencyCount := v.ProvisionedConcurrency

	if provisionedConcurrencyCount > 0 {
		// Provisioned Concurrency Readiness (Warm slots)
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
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ProvisionedConcurrency:Mem-%vGB/", memorySizeMB/1024))},
				},
			},
		})

		// Provisioned Concurrency Execution (Billed when request hits a warm slot)
		provisionedConcurrencyUsage := floatPtrToDecimalPtr(v.MonthlyProvisionedConcurrencyInferenceDurationSeconds)
		components = append(components, &schema.CostComponent{
			Name:            "Provisioned concurrency execution",
			Unit:            "seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: provisionedConcurrencyUsage,
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ProvisionedConcurrency:Usage-%vGB/", memorySizeMB/1024))},
				},
			},
		})
	}

	return components
}

func (s *SageMakerEndpointConfiguration) sagemakerInstanceComponents(variant *SageMakerVariant) []*schema.CostComponent {
	count := decimal.NewFromInt(variant.InitialInstanceCount)

	if count.IsZero() {
		count = decimal.NewFromInt(1)
	}

	components := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("%s (%s)", variant.Label, variant.InstanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: &count,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(variant.InstanceType)},
				},
			},
		},
	}

	if variant.VolumeSizeInGB > 0 {
		monthlyQty := count.Mul(decimal.NewFromInt(variant.VolumeSizeInGB))
		components = append(components, &schema.CostComponent{
			Name:            fmt.Sprintf("%s storage", variant.Label),
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &monthlyQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{
						Key:        "usagetype",
						ValueRegex: strPtr("/Studio:VolumeUsage.gp3/"), //USE1-Studio:VolumeUsage.gp3
					},
				},
			},
		})
	}

	return components
}
