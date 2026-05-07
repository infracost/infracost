package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SageMakerEndpointConfiguration struct {
	Address  string
	Region   string
	Variants []*SageMakerVariant

	MonthlyInstanceHours                                  *int64 `infracost_usage:"monthly_instance_hours"`
	MonthlyInferenceDurationSeconds                       *int64 `infracost_usage:"monthly_inference_duration_seconds"`
	MonthlyProvisionedConcurrencyUsageSeconds             *int64 `infracost_usage:"monthly_provisioned_concurrency_usage_seconds"`
	MonthlyProvisionedConcurrencyInferenceDurationSeconds *int64 `infracost_usage:"monthly_provisioned_concurrency_inference_duration_seconds"`
	DataProcessedOutGB                                    *int64 `infracost_usage:"monthly_data_processed_out_gb"`
	DataProcessedInGB                                     *int64 `infracost_usage:"monthly_data_processed_in_gb"`
	MonthlyStorageDays                                    *int64 `infracost_usage:"monthly_storage_days"`
}

type SageMakerVariant struct {
	Name string

	InstanceType           string
	InitialInstanceCount   int64
	IsServerless           bool
	IsShadow               bool
	MemorySize             int64
	ProvisionedConcurrency int64
	VolumeSizeInGB         int64
	Label                  string
}

func (s *SageMakerEndpointConfiguration) CoreType() string {
	return "SageMakerEndpointConfiguration"
}

func (s *SageMakerEndpointConfiguration) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_inference_duration_seconds", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_provisioned_concurrency_usage_seconds", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_provisioned_concurrency_inference_duration_seconds", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_data_processed_out_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_data_processed_in_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_storage_days", DefaultValue: 30, ValueType: schema.Int64},
	}
}

func (s *SageMakerEndpointConfiguration) UsageEstimationParams() []schema.UsageParam {
	for _, v := range s.Variants {
		if v.IsServerless {
			return []schema.UsageParam{
				{Key: "memory_size_gb", Value: decimal.NewFromInt(v.MemorySize).Div(decimal.NewFromInt(1024)).String()},
			}
		}
	}
	return nil
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
		UsageSchema:    s.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (s *SageMakerEndpointConfiguration) sagemakerServerlessComponents(v *SageMakerVariant) []*schema.CostComponent {
	var components []*schema.CostComponent

	monthlyInferenceSeconds := decimal.NewFromInt(0)
	if s.MonthlyInferenceDurationSeconds != nil {
		monthlyInferenceSeconds = decimal.NewFromInt(*s.MonthlyInferenceDurationSeconds)
	}

	components = append(components, &schema.CostComponent{
		Name:            fmt.Sprintf("Compute Duration (%d MB)", v.MemorySize),
		Unit:            "seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(monthlyInferenceSeconds),
		UsageBased:      true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(s.Region),
			Service:       strPtr("AmazonSageMaker"),
			ProductFamily: strPtr("ML Serverless"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ServerlessInf:Mem-%vGB/", v.MemorySize/1024))},
			},
		},
	})

	if s.DataProcessedOutGB != nil && *s.DataProcessedOutGB > 0 {
		components = append(components, &schema.CostComponent{
			Name:            "Data processed (out)",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(*s.DataProcessedOutGB)),
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(s.Region),
				Service:    strPtr("AmazonSageMaker"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr("Hosting:OUT")},
					{Key: "operation", Value: strPtr("Invoke-Endpoint")},
					{Key: "usagetype", ValueRegex: strPtr("/Data-Bytes-Out/")},
				},
			},
		})
	}

	if s.DataProcessedInGB != nil && *s.DataProcessedInGB > 0 {
		components = append(components, &schema.CostComponent{
			Name:            "Data processed (in)",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(*s.DataProcessedInGB)),
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(s.Region),
				Service:    strPtr("AmazonSageMaker"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr("Hosting:IN")},
					{Key: "operation", Value: strPtr("Invoke-Endpoint")},
					{Key: "usagetype", ValueRegex: strPtr("/Data-Bytes-In/")},
				},
			},
		})
	}

	provisionedConcurrencyCount := v.ProvisionedConcurrency
	if provisionedConcurrencyCount > 0 {
		provisionedConcurrencyUsageSeconds := decimal.NewFromInt(0)
		if s.MonthlyProvisionedConcurrencyUsageSeconds != nil {
			provisionedConcurrencyUsageSeconds = decimal.NewFromInt(*s.MonthlyProvisionedConcurrencyUsageSeconds)
		}
		warmGBSeconds := decimal.NewFromInt(provisionedConcurrencyCount).
			Mul(provisionedConcurrencyUsageSeconds)
		components = append(components, &schema.CostComponent{
			Name:            "Provisioned concurrency (warm)",
			Unit:            "seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(warmGBSeconds),
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr("ServerlessProvisionedConcurrency-Usage")},
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ProvisionedConcurrency:Usage-%vGB/", v.MemorySize/1024))},
				},
			},
		})

		provisionedConcurrencyInferenceDurationSeconds := decimal.NewFromInt(0)
		if s.MonthlyProvisionedConcurrencyInferenceDurationSeconds != nil {
			provisionedConcurrencyInferenceDurationSeconds = decimal.NewFromInt(*s.MonthlyProvisionedConcurrencyInferenceDurationSeconds)
		}

		memorySizeMB := v.MemorySize
		memorySizeGB := decimal.NewFromInt(memorySizeMB).Div(decimal.NewFromInt(1024))
		provisionedConcurrencyExecutionGBSeconds := provisionedConcurrencyInferenceDurationSeconds.Mul(memorySizeGB)

		components = append(components, &schema.CostComponent{
			Name:            "Provisioned concurrency execution",
			Unit:            "seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(provisionedConcurrencyExecutionGBSeconds),
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("ML Serverless"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr("ServerlessProvisionedConcurrency-Duration")},
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/ProvisionedConcurrency:Mem-%vGB/", v.MemorySize/1024))},
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

	monthlyHours := decimal.NewFromInt(730)
	if s.MonthlyInstanceHours != nil {
		print("Using monthly instance hours from usage data: ", *s.MonthlyInstanceHours)
		monthlyHours = decimal.NewFromInt(*s.MonthlyInstanceHours)
	}

	components := []*schema.CostComponent{
		{
			Name:            fmt.Sprintf("Instance (%s)", variant.InstanceType),
			Unit:            "hours",
			UnitMultiplier:  count,
			MonthlyQuantity: &monthlyHours,
			UsageBased:      true,
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
		storageDays := int64(30)
		if s.MonthlyStorageDays != nil && *s.MonthlyStorageDays > 0 {
			storageDays = *s.MonthlyStorageDays
		}

		monthlyQty := decimal.NewFromInt(variant.VolumeSizeInGB).
			Mul(decimal.NewFromInt(storageDays)).
			Div(decimal.NewFromInt(30))
		components = append(components, &schema.CostComponent{
			Name:            fmt.Sprintf("%s storage", variant.Label),
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &monthlyQty,
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(s.Region),
				Service:       strPtr("AmazonSageMaker"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeType", Value: strPtr("General Purpose-Hosting")},
					{Key: "usagetype", ValueRegex: strPtr("/Host:VolumeUsage.gp2/")},
					{Key: "platoclassificationtype", Value: strPtr("Host:VolumeUsage")},
				},
			},
		})
	}

	if s.DataProcessedOutGB != nil && *s.DataProcessedOutGB > 0 {
		components = append(components, &schema.CostComponent{
			Name:            "Data processed (out)",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(*s.DataProcessedOutGB)),
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(s.Region),
				Service:    strPtr("AmazonSageMaker"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr("Hosting:OUT")},
					{Key: "operation", Value: strPtr("Invoke-Endpoint")},
					{Key: "usagetype", ValueRegex: strPtr("/Data-Bytes-Out/")},
				},
			},
		})
	}

	if s.DataProcessedInGB != nil && *s.DataProcessedInGB > 0 {
		components = append(components, &schema.CostComponent{
			Name:            "Data processed (in)",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(*s.DataProcessedInGB)),
			UsageBased:      true,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(s.Region),
				Service:    strPtr("AmazonSageMaker"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "group", Value: strPtr("Hosting:IN")},
					{Key: "operation", Value: strPtr("Invoke-Endpoint")},
					{Key: "usagetype", ValueRegex: strPtr("/Data-Bytes-In/")},
				},
			},
		})
	}

	return components
}
