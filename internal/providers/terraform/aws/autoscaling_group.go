package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAutoscalingGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_autoscaling_group",
		RFunc: NewAutoscalingGroup,
		ReferenceAttributes: []string{
			"launch_configuration",
			"launch_template.0.id",
			"mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id",
		},
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
		lc := newLaunchConfiguration(launchConfigurationRef[0].Address, launchConfigurationRef[0], region)

		// AutoscalingGroup should show as not supported LaunchConfiguration is not supported
		if lc == nil {
			return nil
		}
		multiplyQuantities(lc, desiredCapacity)
		subResources = append(subResources, lc)
	} else if len(launchTemplateRef) > 0 {
		onDemandCount := desiredCapacity
		spotCount := decimal.Zero
		if launchTemplateRef[0].Get("instance_market_options.0.market_type").String() == "spot" {
			onDemandCount = decimal.Zero
			spotCount = desiredCapacity
		}
		lt := newLaunchTemplate(launchTemplateRef[0].Address, launchTemplateRef[0], region, onDemandCount, spotCount)

		// AutoscalingGroup should show as not supported LaunchTemplate is not supported
		if lt == nil {
			return nil
		}
		subResources = append(subResources, lt)
	} else if len(mixedInstanceLaunchTemplateRef) > 0 {
		mixedInstancesPolicy := d.Get("mixed_instances_policy.0")
		lt := newMixedInstancesAwsLaunchTemplate(mixedInstanceLaunchTemplateRef[0].Address, mixedInstanceLaunchTemplateRef[0], region, desiredCapacity, mixedInstancesPolicy)

		// AutoscalingGroup should show as not supported LaunchTemplate is not supported
		if lt == nil {
			return nil
		}
		subResources = append(subResources, lt)
	}

	return &schema.Resource{
		Name:         d.Address,
		SubResources: subResources,
	}
}

func newLaunchConfiguration(name string, d *schema.ResourceData, region string) *schema.Resource {
	tenancy := "Shared"
	if d.Get("placement_tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", d.Address)
		return nil
	} else if d.Get("placement_tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, newRootBlockDevice(d.Get("root_block_device.0"), region))
	subResources = append(subResources, newEbsBlockDevices(d.Get("ebs_block_device"), region)...)

	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}
	costComponents := []*schema.CostComponent{computeCostComponent(d, purchaseOption, tenancy)}

	if d.Get("ebs_optimized").Bool() {
		costComponents = append(costComponents, ebsOptimizedCostComponent(d))
	}

	// Detailed monitoring is enabled by default for launch configurations
	if d.Get("enable_monitoring").Bool() {
		costComponents = append(costComponents, detailedMonitoringCostComponent(d))
	}

	c := cpuCreditsCostComponent(d)
	if c != nil {
		costComponents = append(costComponents, c)
	}

	return &schema.Resource{
		Name:           name,
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}

func newLaunchTemplate(name string, d *schema.ResourceData, region string, onDemandCount decimal.Decimal, spotCount decimal.Decimal) *schema.Resource {
	tenancy := "Shared"
	if d.Get("placement.0.tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Templates", d.Address)
		return nil
	} else if d.Get("placement.0.tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	totalCount := onDemandCount.Add(spotCount)

	costComponents := make([]*schema.CostComponent, 0)

	if d.Get("ebs_optimized").Bool() {
		c := ebsOptimizedCostComponent(d)
		costComponents = append(costComponents, c)
	}

	if d.Get("elastic_inference_accelerator.0.type").Exists() {
		c := elasticInferenceAcceleratorCostComponent(d)
		costComponents = append(costComponents, c)
	}

	if d.Get("monitoring.0.enabled").Bool() {
		c := detailedMonitoringCostComponent(d)
		costComponents = append(costComponents, c)
	}

	c := cpuCreditsCostComponent(d)
	if c != nil {
		costComponents = append(costComponents, c)
	}

	subResources := make([]*schema.Resource, 0)
	rootBlockDevice := newRootBlockDevice(d.Get("root_block_device.0"), region)
	subResources = append(subResources, rootBlockDevice)

	for i, blockDeviceMappingData := range d.Get("block_device_mappings.#.ebs|@flatten").Array() {
		name := fmt.Sprintf("block_device_mapping[%d]", i)
		ebsBlockDevice := newEbsBlockDevice(name, blockDeviceMappingData, region)
		subResources = append(subResources, ebsBlockDevice)
	}

	r := &schema.Resource{
		Name:           name,
		SubResources:   subResources,
		CostComponents: costComponents,
	}

	multiplyQuantities(r, totalCount)

	if spotCount.GreaterThan(decimal.Zero) {
		c := computeCostComponent(d, "spot", tenancy)
		c.HourlyQuantity = decimalPtr(c.HourlyQuantity.Mul(spotCount))
		r.CostComponents = append([]*schema.CostComponent{c}, r.CostComponents...)
	}

	if onDemandCount.GreaterThan(decimal.Zero) {
		c := computeCostComponent(d, "on_demand", tenancy)
		c.HourlyQuantity = decimalPtr(c.HourlyQuantity.Mul(onDemandCount))
		r.CostComponents = append([]*schema.CostComponent{c}, r.CostComponents...)
	}

	return r
}

func newMixedInstancesAwsLaunchTemplate(name string, d *schema.ResourceData, region string, desiredCapacity decimal.Decimal, mixedInstancePolicyData gjson.Result) *schema.Resource {
	overrideInstanceType, totalCount := getInstanceTypeAndCount(mixedInstancePolicyData, desiredCapacity)
	if overrideInstanceType != "" {
		d.Set("instance_type", overrideInstanceType)
	}

	onDemandCount, spotCount := calculateOnDemandAndSpotCounts(mixedInstancePolicyData, totalCount)

	return newLaunchTemplate(name, d, region, onDemandCount, spotCount)
}

func elasticInferenceAcceleratorCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	deviceType := d.Get("elastic_inference_accelerator.0.type").String()

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Inference accelerator (%s)", deviceType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEI"),
			ProductFamily: strPtr("Elastic Inference"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", deviceType))},
			},
		},
	}
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

		if weightedCapacity.Equals(decimal.Zero) {
			count = decimal.Zero
		} else {
			count = capacity.Div(weightedCapacity).Ceil()
		}
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
			costComponent.HourlyQuantity = decimalPtr(costComponent.HourlyQuantity.Mul(multiplier))
		}
		if costComponent.MonthlyQuantity != nil {
			costComponent.MonthlyQuantity = decimalPtr(costComponent.MonthlyQuantity.Mul(multiplier))
		}
	}

	for _, subResource := range resource.SubResources {
		multiplyQuantities(subResource, multiplier)
	}
}
