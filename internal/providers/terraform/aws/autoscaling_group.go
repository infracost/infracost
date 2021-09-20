package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAutoscalingGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_autoscaling_group",
		RFunc: NewAutoscalingGroup,
		ReferenceAttributes: []string{
			"launch_configuration",
			"launch_template.0.id",
			"launch_template.0.name",
			"mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id",
			"launch_template",
		},
	}
}

func NewAutoscalingGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	a := &aws.AutoscalingGroup{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	capacity := d.Get("desired_capacity").Int()

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
		a.LaunchConfiguration = newLaunchConfiguration(data, u, a.Region, capacity)
	} else if len(launchTemplateRef) > 0 {
		data := launchTemplateRef[0]

		onDemandCount := capacity
		spotCount := int64(0)
		if strings.ToLower(d.Get("instance_market_options.0.market_type").String()) == "spot" {
			onDemandCount = int64(0)
			spotCount = capacity
		}

		a.LaunchTemplate = newLaunchTemplate(data, u, a.Region, onDemandCount, spotCount)
	} else if len(mixedInstanceLaunchTemplateRef) > 0 {
		data := mixedInstanceLaunchTemplateRef[0]
		a.LaunchTemplate = newMixedInstancesLaunchTemplate(data, u, a.Region, capacity, d.Get("mixed_instances_policy.0"))
		return nil
	}

	a.PopulateUsage(u)

	return a.BuildResource()
}

func newLaunchConfiguration(d *schema.ResourceData, u *schema.UsageData, region string, count int64) *aws.LaunchConfiguration {
	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}

	a := &aws.LaunchConfiguration{
		Address:          d.Address,
		Region:           region,
		Count:            count,
		Tenancy:          d.Get("placement_tenancy").String(),
		PurchaseOption:   purchaseOption,
		InstanceType:     d.Get("instance_type").String(),
		EBSOptimized:     d.Get("ebs_optimized").Bool(),
		EnableMonitoring: d.Get("enable_monitoring").Bool(),
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

	a.PopulateUsage(u)

	return a
}

func newLaunchTemplate(d *schema.ResourceData, u *schema.UsageData, region string, onDemandCount, spotCount int64) *aws.LaunchTemplate {
	a := &aws.LaunchTemplate{
		Address:          d.Address,
		Region:           region,
		OnDemandCount:    onDemandCount,
		SpotCount:        spotCount,
		Tenancy:          d.Get("placement.0.tenancy").String(),
		InstanceType:     d.Get("instance_type").String(),
		EBSOptimized:     d.Get("ebs_optimized").Bool(),
		EnableMonitoring: d.Get("monitoring.0.enabled").Bool(),
		CPUCredits:       d.Get("credit_specification.0.cpu_credits").String(),
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

	a.PopulateUsage(u)

	return a
}

func newMixedInstancesLaunchTemplate(d *schema.ResourceData, u *schema.UsageData, region string, capacity int64, mixedInstancePolicyData gjson.Result) *aws.LaunchTemplate {
	overrideInstanceType, totalCount := getInstanceTypeAndCount(mixedInstancePolicyData, capacity)
	if overrideInstanceType != "" {
		d.Set("instance_type", overrideInstanceType)
	}

	onDemandCount, spotCount := calculateOnDemandAndSpotCounts(mixedInstancePolicyData, totalCount)

	return newLaunchTemplate(d, u, region, onDemandCount, spotCount)
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

func calculateOnDemandAndSpotCounts(mixedInstancePolicyData gjson.Result, totalCount int64) (int64, int64) {
	instanceDistribution := mixedInstancePolicyData.Get("instances_distribution.0")
	onDemandBaseCount := int64(0)
	if instanceDistribution.Get("on_demand_base_capacity").Exists() {
		onDemandBaseCount = instanceDistribution.Get("on_demand_base_capacity").Int()
	}

	onDemandPerc := int64(100)
	if instanceDistribution.Get("on_demand_percentage_above_base_capacity").Exists() {
		onDemandPerc = instanceDistribution.Get("on_demand_percentage_above_base_capacity").Int()
	}

	onDemandCount := onDemandBaseCount
	remainingCount := totalCount - onDemandCount
	onDemandCount += (remainingCount * decimal.NewFromInt(onDemandPerc).Div(decimal.NewFromInt(100)).Ceil().IntPart())
	spotCount := totalCount - onDemandCount

	return onDemandCount, spotCount
}
