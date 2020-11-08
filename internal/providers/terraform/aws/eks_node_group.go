package aws

import (
	"fmt"

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
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, eksComputeCostComponent(d))
	costComponents = append(costComponents, newEksRootBlockDevice(d))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func eksComputeCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	scalingConfig := d.Get("scaling_config").Array()[0]
	desiredSize := int(scalingConfig.Get("desired_size").Int())
	instanceType := "t3.medium"
	if d.Get("instance_type").Exists() {
		instanceType = d.Get("instance_type").String()
	}
	purchaseOptionLabel := "on-demand"

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Linux/UNIX usage (%s, %s)", purchaseOptionLabel, instanceType),
		Unit:           "hours",
		UnitMultiplier: desiredSize,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(instanceType)},
				{Key: "operatingSystem", Value: strPtr("Linux")},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
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
