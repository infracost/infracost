package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/aws"

	"github.com/infracost/infracost/internal/schema"
)

func GetAutoscalingGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_autoscaling_group",
		CoreRFunc: NewAutoscalingGroup,
		ReferenceAttributes: []string{
			"launch_configuration",
			"launch_template.0.id",
			"launch_template.0.name",
			"mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id",
			"launch_template",
		},
	}
}

func NewAutoscalingGroup(d *schema.ResourceData) schema.CoreResource {
	a := &aws.AutoscalingGroup{
		Address: d.Address,
		Region:  d.Get("region").String(),
		Name:    d.Get("name").String(),
	}

	var instanceCount int64

	if !d.IsEmpty("desired_capacity") {
		instanceCount = d.Get("desired_capacity").Int()
	} else {
		instanceCount = d.Get("min_size").Int()
		if instanceCount == 0 {
			logging.Logger.Debug().Msgf("Using instance count 1 for %s since no desired_capacity or non-zero min_size is set. To override this set the instance_count attribute for this resource in the Infracost usage file.", a.Address)
			instanceCount = 1
		}
	}

	// The Autoscaling Group resource has either a Launch Configuration or Launch Template sub-resource.
	// So we create generic resources for these and add them as a subresource of the Autoscaling Group resource.
	launchConfigurationRef := d.References("launch_configuration")
	launchTemplateRef := d.References("launch_template")
	if len(launchTemplateRef) == 0 {
		launchTemplateRef = d.References("launch_template.0.id")
	}
	if len(launchTemplateRef) == 0 {
		launchTemplateRef = d.References("launch_template.0.name")
	}
	mixedInstanceLaunchTemplateRef := d.References("mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id")

	if len(launchConfigurationRef) > 0 {
		data := launchConfigurationRef[0]
		a.LaunchConfiguration = newLaunchConfiguration(data, a.Region, instanceCount)
	} else if len(launchTemplateRef) > 0 {
		data := launchTemplateRef[0]

		onDemandPercentageAboveBaseCount := int64(100)
		if strings.ToLower(data.Get("instance_market_options.0.market_type").String()) == "spot" {
			onDemandPercentageAboveBaseCount = int64(0)
		}

		a.LaunchTemplate = newLaunchTemplate(data, a.Region, instanceCount, int64(0), onDemandPercentageAboveBaseCount)
	} else if len(mixedInstanceLaunchTemplateRef) > 0 {
		data := mixedInstanceLaunchTemplateRef[0]
		a.LaunchTemplate = newMixedInstancesLaunchTemplate(data, a.Region, instanceCount, d.Get("mixed_instances_policy.0"))
	}

	return a
}

func newLaunchConfiguration(d *schema.ResourceData, region string, instanceCount int64) *aws.LaunchConfiguration {
	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}

	a := &aws.LaunchConfiguration{
		Address:          d.Address,
		Region:           region,
		AMI:              d.Get("image_id").String(),
		InstanceCount:    intPtr(instanceCount),
		Tenancy:          d.Get("placement_tenancy").String(),
		PurchaseOption:   purchaseOption,
		InstanceType:     d.Get("instance_type").String(),
		EBSOptimized:     d.Get("ebs_optimized").Bool(),
		EnableMonitoring: d.GetBoolOrDefault("enable_monitoring", true),
		CPUCredits:       d.Get("credit_specification.0.cpu_credits").String(),
	}

	a.RootBlockDevice = &aws.EBSVolume{
		Address: "root_block_device",
		Region:  region,
		Type:    d.Get("root_block_device.0.volume_type").String(),
		IOPS:    d.Get("root_block_device.0.iops").Int(),
	}

	if d.Get("root_block_device.0.volume_size").Type != gjson.Null {
		a.RootBlockDevice.Size = intPtr(d.Get("root_block_device.0.volume_size").Int())
	}

	for i, data := range d.Get("ebs_block_device").Array() {
		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("ebs_block_device[%d]", i),
			Region:  region,
			Type:    data.Get("volume_type").String(),
			IOPS:    data.Get("iops").Int(),
		}

		if data.Get("volume_size").Type != gjson.Null {
			ebsBlockDevice.Size = intPtr(data.Get("volume_size").Int())
		}

		a.EBSBlockDevices = append(a.EBSBlockDevices, ebsBlockDevice)
	}

	return a
}

func newLaunchTemplate(d *schema.ResourceData, region string, instanceCount, onDemandBaseCount, onDemandPercentageAboveBaseCount int64) *aws.LaunchTemplate {
	a := &aws.LaunchTemplate{
		Address:                          d.Address,
		Region:                           region,
		AMI:                              d.Get("image_id").String(),
		InstanceCount:                    intPtr(instanceCount),
		OnDemandBaseCount:                onDemandBaseCount,
		OnDemandPercentageAboveBaseCount: onDemandPercentageAboveBaseCount,
		Tenancy:                          d.Get("placement.0.tenancy").String(),
		InstanceType:                     d.Get("instance_type").String(),
		EBSOptimized:                     d.Get("ebs_optimized").Bool(),
		EnableMonitoring:                 d.Get("monitoring.0.enabled").Bool(),
		CPUCredits:                       d.Get("credit_specification.0.cpu_credits").String(),
	}

	if d.Get("elastic_inference_accelerator.0.type").Type != gjson.Null {
		a.ElasticInferenceAcceleratorType = strPtr(d.Get("elastic_inference_accelerator.0.type").String())
	}

	for i, data := range d.Get("block_device_mappings.#.ebs|@flatten").Array() {
		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("block_device_mapping[%d]", i),
			Region:  region,
			Type:    data.Get("volume_type").String(),
			IOPS:    data.Get("iops").Int(),
		}

		if data.Get("volume_size").Type != gjson.Null {
			ebsBlockDevice.Size = intPtr(data.Get("volume_size").Int())
		}

		a.EBSBlockDevices = append(a.EBSBlockDevices, ebsBlockDevice)
	}

	return a
}

func newMixedInstancesLaunchTemplate(d *schema.ResourceData, region string, capacity int64, mixedInstancePolicyData gjson.Result) *aws.LaunchTemplate {
	overrideInstanceType, instanceCount := getInstanceTypeAndCount(mixedInstancePolicyData, capacity)
	if overrideInstanceType != "" {
		d.Set("instance_type", overrideInstanceType)
	}

	instanceDistribution := mixedInstancePolicyData.Get("instances_distribution.0")
	onDemandBaseCount := int64(0)
	if instanceDistribution.Get("on_demand_base_capacity").Exists() {
		onDemandBaseCount = instanceDistribution.Get("on_demand_base_capacity").Int()
	}

	onDemandPercentageAboveBaseCount := int64(100)
	if instanceDistribution.Get("on_demand_percentage_above_base_capacity").Exists() {
		onDemandPercentageAboveBaseCount = instanceDistribution.Get("on_demand_percentage_above_base_capacity").Int()
	}

	return newLaunchTemplate(d, region, instanceCount, onDemandBaseCount, onDemandPercentageAboveBaseCount)
}

func getInstanceTypeAndCount(mixedInstancePolicyData gjson.Result, capacity int64) (string, int64) {
	count := capacity
	instanceType := ""

	override := mixedInstancePolicyData.Get("launch_template.0.override.0")
	if override.Exists() {
		instanceType = override.Get("instance_type").String()
		weightedCapacity := int64(1)
		if override.Get("weighted_capacity").Type != gjson.Null {
			weightedCapacity = override.Get("weighted_capacity").Int()
		}

		if weightedCapacity == 0 {
			count = int64(0)
		} else {
			count = decimal.NewFromInt(capacity).Div(decimal.NewFromInt(weightedCapacity)).Ceil().IntPart()
		}
	}

	return instanceType, count
}
