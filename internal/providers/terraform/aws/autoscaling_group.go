package aws

import (
	"fmt"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAutoscalingGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_autoscaling_group",
		RFunc: NewAutoscalingGroup,
	}
}

func GetLaunchConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:   "aws_launch_configuration",
		NoCost: true,
	}
}

func GetLaunchTemplateRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:   "aws_launch_template",
		NoCost: true,
	}
}

func NewAutoscalingGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()
	desiredCapacity := decimal.NewFromInt(d.Get("desired_capacity").Int())

	subResources := make([]*schema.Resource, 0)

	launchConfigurationRef := d.References("launch_configuration")
	launchTemplateRef := d.References("launch_template.0.id")
	mixedInstanceLaunchTemplateRef := d.References("mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id")

	if len(launchConfigurationRef) > 0 {
		launchConfiguration := newLaunchConfiguration(launchConfigurationRef[0].Address, launchConfigurationRef[0], region)
		multiplyQuantities(launchConfiguration, desiredCapacity)
		subResources = append(subResources, launchConfiguration)
	} else if len(launchTemplateRef) > 0 {
		subResources = append(subResources, newLaunchTemplate(launchTemplateRef[0].Address, launchTemplateRef[0], region, desiredCapacity, decimal.Zero))
	} else if len(mixedInstanceLaunchTemplateRef) > 0 {
		mixedInstancesPolicy := d.Get("mixed_instances_policy.0")
		subResources = append(subResources, newMixedInstancesAwsLaunchTemplate(mixedInstanceLaunchTemplateRef[0].Address, mixedInstanceLaunchTemplateRef[0], region, desiredCapacity, mixedInstancesPolicy))
	}

	return &schema.Resource{
		Name:         d.Address,
		SubResources: subResources,
	}
}

func newLaunchConfiguration(name string, d *schema.ResourceData, region string) *schema.Resource {
	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, newRootBlockDevice(d.Get("root_block_device.0"), region))
	subResources = append(subResources, newEbsBlockDevices(d.Get("ebs_block_device"), region)...)

	return &schema.Resource{
		Name:           name,
		SubResources:   subResources,
		CostComponents: computeCostComponents(d, region, "on_demand"),
	}
}

func newLaunchTemplate(name string, d *schema.ResourceData, region string, onDemandCount decimal.Decimal, spotCount decimal.Decimal) *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	if onDemandCount.GreaterThan(decimal.Zero) {
		cs := computeCostComponents(d, region, "on_demand")
		for _, c := range cs {
			c.HourlyQuantity = decimalPtr(c.HourlyQuantity.Mul(onDemandCount))
		}
		costComponents = append(costComponents, cs...)
	}

	if spotCount.GreaterThan(decimal.Zero) {
		cs := computeCostComponents(d, region, "spot")
		for _, c := range cs {
			c.HourlyQuantity = decimalPtr(c.HourlyQuantity.Mul(spotCount))
		}
		costComponents = append(costComponents, cs...)
	}

	subResources := make([]*schema.Resource, 0)

	totalCount := onDemandCount.Add(spotCount)
	rootBlockDevice := newRootBlockDevice(d.Get("root_block_device.0"), region)
	multiplyQuantities(rootBlockDevice, totalCount)
	subResources = append(subResources, rootBlockDevice)

	for i, blockDeviceMappingData := range d.Get("block_device_mappings.#.ebs|@flatten").Array() {
		name := fmt.Sprintf("block_device_mapping[%d]", i)
		ebsBlockDevice := newEbsBlockDevice(name, blockDeviceMappingData, region)
		multiplyQuantities(ebsBlockDevice, totalCount)
		subResources = append(subResources, ebsBlockDevice)
	}

	return &schema.Resource{
		Name:           name,
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}

func newMixedInstancesAwsLaunchTemplate(name string, d *schema.ResourceData, region string, desiredCapacity decimal.Decimal, mixedInstancePolicyData gjson.Result) *schema.Resource {
	overrideInstanceType, totalCount := getInstanceTypeAndCount(mixedInstancePolicyData, desiredCapacity)
	if overrideInstanceType != "" {
		d.Set("instance_type", overrideInstanceType)
	}

	onDemandCount, spotCount := calculateOnDemandAndSpotCounts(mixedInstancePolicyData, totalCount)

	return newLaunchTemplate(name, d, region, onDemandCount, spotCount)
}

func getInstanceTypeAndCount(mixedInstancePolicyData gjson.Result, capacity decimal.Decimal) (string, decimal.Decimal) {
	count := capacity
	instanceType := ""

	override := mixedInstancePolicyData.Get("launch_template.0.override.0")
	if override.Exists() {
		instanceType = override.Get("instance_type").String()
		weightedCapacity := decimal.NewFromInt(1)
		if override.Get("weighted_capacity").Exists() && override.Get("weighted_capacity").Type != gjson.Null {
			weightedCapacity = decimal.NewFromInt(override.Get("weighted_capacity").Int())
		}
		count = capacity.Div(weightedCapacity).Ceil()
	}

	return instanceType, count
}

func calculateOnDemandAndSpotCounts(mixedInstancePolicyData gjson.Result, totalCount decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	instanceDistribution := mixedInstancePolicyData.Get("instances_distribution.0")
	onDemandBaseCount := decimal.Zero
	if instanceDistribution.Get("on_demand_base_capacity").Exists() {
		onDemandBaseCount = decimal.NewFromInt(instanceDistribution.Get("on_demand_base_capacity").Int())
	}

	onDemandPerc := decimal.NewFromInt(100)
	if instanceDistribution.Get("on_demand_percentage_above_base_capacity").Exists() {
		onDemandPerc = decimal.NewFromInt(instanceDistribution.Get("on_demand_percentage_above_base_capacity").Int())
	}

	onDemandCount := onDemandBaseCount
	remainingCount := totalCount.Sub(onDemandCount)
	onDemandCount = onDemandCount.Add(remainingCount.Mul(onDemandPerc).Div(decimal.NewFromInt(100)).Ceil())
	spotCount := totalCount.Sub(onDemandCount)

	return onDemandCount, spotCount
}

func multiplyQuantities(resource *schema.Resource, multiplier decimal.Decimal) {
	for _, costComponent := range resource.CostComponents {
		if costComponent.HourlyQuantity != nil {
			costComponent.HourlyQuantity = decimalPtr((*costComponent.HourlyQuantity).Mul(multiplier))
		}
		if costComponent.MonthlyQuantity != nil {
			costComponent.MonthlyQuantity = decimalPtr((*costComponent.MonthlyQuantity).Mul(multiplier))
		}
	}

	for _, subResource := range resource.SubResources {
		multiplyQuantities(subResource, multiplier)
	}
}
