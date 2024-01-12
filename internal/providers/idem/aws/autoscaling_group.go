package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/aws"

	"github.com/infracost/infracost/internal/schema"
)

func GetAutoscalingGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.autoscaling.auto_scaling_group.present",
		RFunc: NewAutoscalingGroup,
		ReferenceAttributes: []string{
			"states.aws.autoscaling.auto_scaling_group.present:launch_configuration_name",
			"states.aws.ec2.launch_template.present:launch_template.LaunchTemplateId",
			"states.aws.ec2.launch_template.present:mixed_instances_policy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId",
		},
	}
}

func NewAutoscalingGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
			log.Debugf("Using instance count 1 for %s since no desired_capacity or non-zero min_size is set. To override this set the instance_count attribute for this resource in the Infracost usage file.", a.Address)
			instanceCount = 1
		}
	}

	// The Autoscaling Group resource has either a Launch Configuration or Launch Template sub-resource.
	// So we create generic resources for these and add them as a subresource of the Autoscaling Group resource.
	launchConfigurationRef := d.References("states.aws.autoscaling.auto_scaling_group.present:launch_configuration_name")
	launchTemplateRef := d.References("states.aws.ec2.launch_template.present:launch_template.LaunchTemplateId")
	mixedInstanceLaunchTemplateRef := d.References("states.aws.ec2.launch_template.present:mixed_instances_policy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId")

	if len(launchConfigurationRef) > 0 {
		data := launchConfigurationRef[0]
		a.LaunchConfiguration = newLaunchConfiguration(data, u, a.Region, instanceCount)
	} else if len(launchTemplateRef) > 0 {
		data := launchTemplateRef[0]

		onDemandPercentageAboveBaseCount := int64(100)
		if strings.ToLower(data.Get("InstanceMarketOptions.MarketType").String()) == "spot" {
			onDemandPercentageAboveBaseCount = int64(0)
		}

		a.LaunchTemplate = newLaunchTemplate(data, a.Region, instanceCount, int64(0), onDemandPercentageAboveBaseCount)
	} else if len(mixedInstanceLaunchTemplateRef) > 0 {
		data := mixedInstanceLaunchTemplateRef[0]
		a.LaunchTemplate = newMixedInstancesLaunchTemplate(data, u, a.Region, instanceCount, d.Get("mixed_instances_policy"))
	}

	a.PopulateUsage(u)

	return a.BuildResource()
}

func newLaunchConfiguration(d *schema.ResourceData, u *schema.UsageData, region string, instanceCount int64) *aws.LaunchConfiguration {
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
		EnableMonitoring: d.GetBoolOrDefault("instance_monitoring.Enabled", true),
		// TODO
		//CPUCredits:       d.Get("launch_template_data.CreditSpecification.CpuCredits").String(),
	}

	/*
		// TODO: What is root?
		a.RootBlockDevice = &aws.EBSVolume{
			Address: "root_block_device",
			Region:  region,
			Type:    d.Get("launch_template_data.root_block_device.BlockDeviceMappings.0.Ebs.VolumeType").String(),
			IOPS:    d.Get("launch_template_data.root_block_device.BlockDeviceMappings.0.Ebs.Iops").Int(),
		}

		if d.Get("launch_template_data.root_block_device.BlockDeviceMappings.0.Ebs.VolumeSize").Type != gjson.Null {
			a.RootBlockDevice.Size = intPtr(d.Get("launch_template_data.root_block_device.BlockDeviceMappings.0.Ebs.VolumeSize").Int())
		}*/

	for i, data := range d.Get("launch_template_data.BlockDeviceMappings").Array() {
		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("ebs_block_device[%d]", i),
			Region:  region,
			Type:    data.Get("VolumeSize").String(),
			IOPS:    data.Get("Iops").Int(),
		}

		if data.Get("VolumeSize").Type != gjson.Null {
			ebsBlockDevice.Size = intPtr(data.Get("VolumeSize").Int())
		}

		a.EBSBlockDevices = append(a.EBSBlockDevices, ebsBlockDevice)
	}

	return a
}

func newLaunchTemplate(d *schema.ResourceData, region string, instanceCount, onDemandBaseCount, onDemandPercentageAboveBaseCount int64) *aws.LaunchTemplate {
	a := &aws.LaunchTemplate{
		Address:                          d.Address,
		Region:                           region,
		AMI:                              d.Get("launch_template_data.ImageId").String(),
		InstanceCount:                    intPtr(instanceCount),
		OnDemandBaseCount:                onDemandBaseCount,
		OnDemandPercentageAboveBaseCount: onDemandPercentageAboveBaseCount,
		Tenancy:                          d.Get("launch_template_data.Placement.Tenancy").String(),
		InstanceType:                     d.Get("launch_template_data.InstanceType").String(),
		EBSOptimized:                     d.Get("launch_template_data.EbsOptimized").Bool(),
		EnableMonitoring:                 d.Get("launch_template_data.Monitoring.Enabled").Bool(),
		CPUCredits:                       d.Get("CreditSpecification.0.cpu_credits").String(),
	}

	if d.Get("launch_template_data.ElasticInferenceAccelerators.0.Type").Type != gjson.Null {
		a.ElasticInferenceAcceleratorType = strPtr(d.Get("launch_template_data.ElasticInferenceAccelerators.0.Type").String())
	}

	for i, data := range d.Get("launch_template_data.BlockDeviceMappings.#.Ebs|@flatten").Array() {
		ebsBlockDevice := &aws.EBSVolume{
			Address: fmt.Sprintf("block_device_mapping[%d]", i),
			Region:  region,
			Type:    data.Get("VolumeType").String(),
			IOPS:    data.Get("Iops").Int(),
		}

		if data.Get("VolumeSize").Type != gjson.Null {
			ebsBlockDevice.Size = intPtr(data.Get("VolumeSize").Int())
		}

		a.EBSBlockDevices = append(a.EBSBlockDevices, ebsBlockDevice)
	}

	return a
}

func newMixedInstancesLaunchTemplate(d *schema.ResourceData, u *schema.UsageData, region string, capacity int64, mixedInstancePolicyData gjson.Result) *aws.LaunchTemplate {
	overrideInstanceType, instanceCount := getInstanceTypeAndCount(mixedInstancePolicyData, capacity)
	if overrideInstanceType != "" {
		d.Set("launch_template_data.InstanceType", overrideInstanceType)
	}

	instanceDistribution := mixedInstancePolicyData.Get("InstancesDistribution")
	onDemandBaseCount := int64(0)
	if instanceDistribution.Get("OnDemandBaseCapacity").Exists() {
		onDemandBaseCount = instanceDistribution.Get("OnDemandBaseCapacity").Int()
	}

	onDemandPercentageAboveBaseCount := int64(100)
	if instanceDistribution.Get("OnDemandPercentageAboveBaseCapacity").Exists() {
		onDemandPercentageAboveBaseCount = instanceDistribution.Get("OnDemandPercentageAboveBaseCapacity").Int()
	}

	return newLaunchTemplate(d, region, instanceCount, onDemandBaseCount, onDemandPercentageAboveBaseCount)
}

func getInstanceTypeAndCount(mixedInstancePolicyData gjson.Result, capacity int64) (string, int64) {
	count := capacity
	instanceType := ""

	override := mixedInstancePolicyData.Get("LaunchTemplate.Overrides.0")
	if override.Exists() {
		instanceType = override.Get("InstanceType").String()
		weightedCapacity := int64(1)
		if override.Get("WeightedCapacity").Type != gjson.Null {
			weightedCapacity = override.Get("WeightedCapacity").Int()
		}

		if weightedCapacity == 0 {
			count = int64(0)
		} else {
			count = decimal.NewFromInt(capacity).Div(decimal.NewFromInt(weightedCapacity)).Ceil().IntPart()
		}
	}

	return instanceType, count
}
