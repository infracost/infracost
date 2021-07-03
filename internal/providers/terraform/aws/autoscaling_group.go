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
			"launch_template.0.name",
			"mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id",
		},
	}
}

func NewAutoscalingGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	desiredCapacity := decimal.NewFromInt(d.Get("desired_capacity").Int())
	if u != nil && u.Get("instances").Exists() {
		if desiredCapacity.GreaterThan(decimal.Zero) {
			log.Debugf("Overriding the desired_capacity for %s by usage data", d.Address)
		}
		desiredCapacity = decimal.NewFromInt(u.Get("instances").Int())
	}

	subResources := make([]*schema.Resource, 0)

	launchConfigurationRef := d.References("launch_configuration")
	launchTemplateRefID := d.References("launch_template.0.id")
	launchTemplateRefName := d.References("launch_template.0.name")
	launchTemplateRef := []*schema.ResourceData{}
	if len(launchTemplateRefID) > 0 {
		launchTemplateRef = launchTemplateRefID
	} else if len(launchTemplateRefName) > 0 {
		launchTemplateRef = launchTemplateRefName
	}
	mixedInstanceLaunchTemplateRef := d.References("mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id")

	if len(launchConfigurationRef) > 0 {
		lc := newLaunchConfiguration(launchConfigurationRef[0].Address, launchConfigurationRef[0], u, region)

		// AutoscalingGroup should show as not supported LaunchConfiguration is not supported
		if lc == nil {
			return nil
		}
		schema.MultiplyQuantities(lc, desiredCapacity)
		subResources = append(subResources, lc)
	} else if len(launchTemplateRef) > 0 {
		onDemandCount := desiredCapacity
		spotCount := decimal.Zero
		if launchTemplateRef[0].Get("instance_market_options.0.market_type").String() == "spot" {
			onDemandCount = decimal.Zero
			spotCount = desiredCapacity
		}
		lt := newLaunchTemplate(launchTemplateRef[0].Address, launchTemplateRef[0], u, region, onDemandCount, spotCount)

		// AutoscalingGroup should show as not supported LaunchTemplate is not supported
		if lt == nil {
			return nil
		}
		subResources = append(subResources, lt)
	} else if len(mixedInstanceLaunchTemplateRef) > 0 {
		mixedInstancesPolicy := d.Get("mixed_instances_policy.0")
		lt := newMixedInstancesAwsLaunchTemplate(mixedInstanceLaunchTemplateRef[0].Address, mixedInstanceLaunchTemplateRef[0], u, region, desiredCapacity, mixedInstancesPolicy)

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

func newLaunchConfiguration(name string, d *schema.ResourceData, u *schema.UsageData, region string) *schema.Resource {
	tenancy := "Shared"
	if d.Get("placement_tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", d.Address)
		return nil
	} else if d.Get("placement_tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	instanceType := d.Get("instance_type").String()
	ami := d.Get("image_id").String()

	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, newRootBlockDevice(d.Get("root_block_device.0"), region))
	subResources = append(subResources, newEbsBlockDevices(d.Get("ebs_block_device"), region)...)

	purchaseOption := "on_demand"
	if d.Get("spot_price").String() != "" {
		purchaseOption = "spot"
	}
	costComponents := []*schema.CostComponent{computeCostComponent(d, u, purchaseOption, instanceType, ami, tenancy, 1)}

	if d.Get("ebs_optimized").Bool() {
		costComponents = append(costComponents, ebsOptimizedCostComponent(d))
	}

	// Detailed monitoring is enabled by default for launch configurations
	if d.Get("enable_monitoring").Bool() {
		costComponents = append(costComponents, detailedMonitoringCostComponent(d))
	}

	c := newCPUCredit(d, u)
	if c != nil {
		costComponents = append(costComponents, c)
	}

	return &schema.Resource{
		Name:           name,
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}

func newLaunchTemplate(name string, d *schema.ResourceData, u *schema.UsageData, region string, onDemandCount decimal.Decimal, spotCount decimal.Decimal) *schema.Resource {
	tenancy := "Shared"
	if d.Get("placement.0.tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Templates", d.Address)
		return nil
	} else if d.Get("placement.0.tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	instanceType := d.Get("instance_type").String()
	ami := d.Get("image_id").String()

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

	if d.Get("instance_type").Exists() && d.Get("instance_type").Type != gjson.Null {
		if isInstanceBurstable(d.Get("instance_type").String(), []string{"t2.", "t3.", "t4."}) {
			c := newCPUCredit(d, u)
			if c != nil {
				costComponents = append(costComponents, c)
			}
		}
	}

	subResources := make([]*schema.Resource, 0)

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

	schema.MultiplyQuantities(r, totalCount)

	if spotCount.GreaterThan(decimal.Zero) {
		c := computeCostComponent(d, u, "spot", instanceType, ami, tenancy, 1)
		c.HourlyQuantity = decimalPtr(c.HourlyQuantity.Mul(spotCount))
		r.CostComponents = append([]*schema.CostComponent{c}, r.CostComponents...)
	}

	if onDemandCount.GreaterThan(decimal.Zero) {
		c := computeCostComponent(d, u, "on_demand", instanceType, ami, tenancy, 1)
		c.HourlyQuantity = decimalPtr(c.HourlyQuantity.Mul(onDemandCount))
		r.CostComponents = append([]*schema.CostComponent{c}, r.CostComponents...)
	}

	return r
}

func newMixedInstancesAwsLaunchTemplate(name string, d *schema.ResourceData, u *schema.UsageData, region string, desiredCapacity decimal.Decimal, mixedInstancePolicyData gjson.Result) *schema.Resource {
	overrideInstanceType, totalCount := getInstanceTypeAndCount(mixedInstancePolicyData, desiredCapacity)
	if overrideInstanceType != "" {
		d.Set("instance_type", overrideInstanceType)
	}

	onDemandCount, spotCount := calculateOnDemandAndSpotCounts(mixedInstancePolicyData, totalCount)

	return newLaunchTemplate(name, d, u, region, onDemandCount, spotCount)
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
