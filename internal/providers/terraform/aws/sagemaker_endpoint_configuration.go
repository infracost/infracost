package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getSagemakerEndpointConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sagemaker_endpoint_configuration",
		RFunc: newSageMakerEndpointConfiguration,
	}
}

func newSageMakerEndpointConfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	s := &aws.SageMakerEndpointConfiguration{
		Address: d.Address,
		Region:  region,
	}

	if d.Get("production_variants").Exists() {
		for _, variant := range d.Get("production_variants").Array() {
			s.Variants = append(s.Variants, decodeVariant(variant, "Inference instance"))
		}
	}

	if d.Get("shadow_production_variants").Exists() {
		for _, variant := range d.Get("shadow_production_variants").Array() {
			s.Variants = append(s.Variants, decodeVariant(variant, "Shadow instance"))
		}
	}
	s.PopulateUsage(u)

	return s.BuildResource()
}

func decodeVariant(v gjson.Result, label string) *aws.SageMakerVariant {
	serverlessConfig := v.Get("serverless_config").Array()
	isServerless := len(serverlessConfig) > 0

	variant := &aws.SageMakerVariant{
		Name:                 v.Get("variant_name").String(),
		InstanceType:         v.Get("instance_type").String(),
		InitialInstanceCount: v.Get("initial_instance_count").Int(),
		VolumeSizeInGB:       v.Get("volume_size_in_gb").Int(),
		Label:                label,
	}

	if isServerless {
		variant.IsServerless = true
		variant.MemorySizeMB = serverlessConfig[0].Get("memory_size_in_mb").Int()
		variant.ProvisionedConcurrency = serverlessConfig[0].Get("provisioned_concurrency").Int()
		variant.MaxConcurrency = serverlessConfig[0].Get("max_concurrency").Int()
	}

	return variant
}
