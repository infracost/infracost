package aws

import (
	"github.com/tidwall/gjson"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewEKSNodeGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_node_group",
		RFunc: NewEKSNodeGroup,
		ReferenceAttributes: []string{
			"launch_template.0.id",
			"launch_template.0.name",
		},
	}
}

func NewEKSNodeGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	scalingConfig := d.Get("scaling_config").Array()[0]
	desiredSize := scalingConfig.Get("desired_size").Int()
	purchaseOptionLabel := "on_demand"
	if d.Get("capacity_type").String() != "" {
		purchaseOptionLabel = strings.ToLower(d.Get("capacity_type").String())
	}
	instanceType := "t3.medium"

	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	launchTemplateRefID := d.References("launch_template.0.id")
	launchTemplateRefName := d.References("launch_template.0.name")
	launchTemplateRef := []*schema.ResourceData{}

	if len(launchTemplateRefID) > 0 {
		launchTemplateRef = launchTemplateRefID
	} else if len(launchTemplateRefName) > 0 {
		launchTemplateRef = launchTemplateRefName
	}

	if len(d.Get("instance_types").Array()) > 0 && len(launchTemplateRef) < 1 || len(launchTemplateRef) < 1 {
		if len(d.Get("instance_types").Array()) > 0 {
			instanceType = strings.ToLower(d.Get("instance_types").Array()[0].String())
		}

		costComponents = append(costComponents, computeCostComponent(d, u, purchaseOptionLabel, instanceType, "Shared", desiredSize))

		var cpuCreditQuantity decimal.Decimal
		if isInstanceBurstable(instanceType, []string{"t3", "t4"}) {
			instanceCPUCreditHours := decimal.Zero
			if u != nil && u.Get("cpu_credit_hrs").Exists() {
				instanceCPUCreditHours = decimal.NewFromInt(u.Get("cpu_credit_hrs").Int())
			}

			instanceVCPUCount := decimal.Zero
			if u != nil && u.Get("virtual_cpu_count").Exists() {
				instanceVCPUCount = decimal.NewFromInt(u.Get("virtual_cpu_count").Int())
			}

			cpuCreditQuantity = instanceVCPUCount.Mul(instanceCPUCreditHours).Mul(decimal.NewFromInt(desiredSize))
			instancePrefix := strings.SplitN(instanceType, ".", 2)[0]
			costComponents = append(costComponents, cpuCreditsCostComponent(region, cpuCreditQuantity, instancePrefix))
		}

		costComponents = append(costComponents, newEksRootBlockDevice(d))
	}

	if len(launchTemplateRef) > 0 {
		spotCount := decimal.Zero
		onDemandCount := decimal.NewFromInt(desiredSize)

		if launchTemplateRef[0].Get("instance_market_options.0.market_type").String() == "spot" {
			onDemandCount = decimal.Zero
			spotCount = decimal.NewFromInt(desiredSize)
		}

		if launchTemplateRef[0].Get("instance_type").Type == gjson.Null {
			launchTemplateRef[0].Set("instance_type", d.Get("instance_types").Array()[0].String())
		}

		lt := newLaunchTemplate(launchTemplateRef[0].Address, launchTemplateRef[0], u, region, onDemandCount, spotCount)

		// AutoscalingGroup should show as not supported LaunchTemplate is not supported
		if lt == nil {
			return nil
		}
		subResources = append(subResources, lt)
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func newEksRootBlockDevice(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return newEksEbsBlockDevice("root_block_device", d, region)
}

func newEksEbsBlockDevice(name string, d *schema.ResourceData, region string) *schema.CostComponent {
	volumeAPIName := "gp2"
	defaultVolumeSize := 20

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("disk_size").Exists() {
		gbVal = decimal.NewFromFloat(d.Get("disk_size").Float())
	}

	iopsVal := decimal.Zero

	var unknown *decimal.Decimal

	return ebsVolumeCostComponents(region, volumeAPIName, unknown, gbVal, iopsVal, unknown)[0]
}
