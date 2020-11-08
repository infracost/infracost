package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewEKSNodeGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_node_group",
		RFunc: NewEKSNodeGroup,
	}
}

func NewEKSNodeGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()
	scalingConfig := d.Get("scaling_config").Array()[0]
	desiredSize := scalingConfig.Get("desired_size").Int()
	instanceType := "t3.medium"
	if d.Get("instance_types").Exists() {
		// Only a single type is expected https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_node_group#instance_types
		instanceType = d.Get("instance_types").Array()[0].String()
	}

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, eksComputeCostComponent(d, region, desiredSize, instanceType))
	eksCPUCreditsCostComponent := eksCPUCreditsCostComponent(d, region, desiredSize, instanceType)
	if eksCPUCreditsCostComponent != nil {
		costComponents = append(costComponents, eksCPUCreditsCostComponent)
	}
	costComponents = append(costComponents, newEksRootBlockDevice(d))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func eksComputeCostComponent(d *schema.ResourceData, region string, desiredSize int64, instanceType string) *schema.CostComponent {

	purchaseOptionLabel := "on-demand"

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Linux/UNIX usage (%s, %s)", purchaseOptionLabel, instanceType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(desiredSize)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(instanceType)},
				{Key: "operatingSystem", Value: strPtr("Linux")},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "tenancy", Value: strPtr("Shared")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func eksCPUCreditsCostComponent(d *schema.ResourceData, region string, desiredSize int64, instanceType string) *schema.CostComponent {

	if !(strings.HasPrefix(instanceType, "t3.") || strings.HasPrefix(instanceType, "t4g.")) {
		return nil
	}

	prefix := strings.SplitN(instanceType, ".", 2)[0]

	return &schema.CostComponent{
		Name:           "CPU credits",
		Unit:           "vCPU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(desiredSize)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("CPU Credits"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operatingSystem", Value: strPtr("Linux")},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/CPUCredits:%s/", prefix))},
			},
		},
	}
}

func newEksRootBlockDevice(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return newEksEbsBlockDevice("root_block_device", d, region)
}

func newEksEbsBlockDevice(name string, d *schema.ResourceData, region string) *schema.CostComponent {
	volumeAPIName := "gp2"

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("disk_size").Exists() {
		gbVal = decimal.NewFromFloat(d.Get("disk_size").Float())
	}

	iopsVal := decimal.Zero

	return ebsVolumeCostComponents(region, volumeAPIName, gbVal, iopsVal)[0]

}
