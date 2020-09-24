package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/pkg/schema"
	log "github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_instance",
		Notes: []string{"Non-Linux EC2 instances such as Windows and RHEL are not supported, a lookup is needed to find the OS of AMIs."},
		RFunc: NewInstance,
	}
}

func NewInstance(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	if d.Get("tenancy").Exists() && d.Get("tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS EC2 instances", d.Address)
		return nil
	}

	region := d.Get("region").String()
	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, newRootBlockDevice(d.Get("root_block_device.0"), region))
	subResources = append(subResources, newEbsBlockDevices(d.Get("ebs_block_device"), region)...)

	return &schema.Resource{
		Name:           d.Address,
		SubResources:   subResources,
		CostComponents: computeCostComponents(d, region, "on_demand"),
	}
}

func computeCostComponents(d *schema.ResourceData, region string, purchaseOption string) []*schema.CostComponent {
	instanceType := d.Get("instance_type").String()

	tenancy := "Shared"
	if d.Get("tenancy").Exists() && d.Get("tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	purchaseOptionLabel := map[string]string{
		"on_demand": "on-demand",
		"spot":      "spot",
	}[purchaseOption]

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Linux/UNIX Usage (%s, %s)", purchaseOptionLabel, instanceType),
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Compute Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(instanceType)},
					{Key: "tenancy", Value: strPtr(tenancy)},
					{Key: "operatingSystem", Value: strPtr("Linux")},
					{Key: "preInstalledSw", Value: strPtr("NA")},
					{Key: "capacitystatus", Value: strPtr("Used")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: &purchaseOption,
			},
		},
	}

	if d.Get("ebs_optimized").Bool() {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:                 "EBS-Optimized Usage",
			Unit:                 "hours",
			HourlyQuantity:       decimalPtr(decimal.NewFromInt(1)),
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Compute Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(instanceType)},
					{Key: "usagetype", ValueRegex: strPtr("/EBSOptimized/")},
				},
			},
		})
	}

	cpuCredits := d.Get("credit_specification.0.cpu_credits").String()
	if cpuCredits == "" && (strings.HasPrefix(instanceType, "t3.") || strings.HasPrefix(instanceType, "t4g.")) {
		cpuCredits = "unlimited"
	}

	if cpuCredits == "unlimited" {
		prefix := strings.SplitN(instanceType, ".", 2)[0]

		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "CPU Credits",
			Unit:           "vCPU-hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(0)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("CPU Credits"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "operatingSystem", Value: strPtr("Linux")},
					{Key: "usagetype", Value: strPtr(fmt.Sprintf("CPUCredits:%s", prefix))},
				},
			},
		})
	}

	return costComponents
}

func newRootBlockDevice(d gjson.Result, region string) *schema.Resource {
	return newEbsBlockDevice("root_block_device", d, region)
}

func newEbsBlockDevices(d gjson.Result, region string) []*schema.Resource {
	resources := make([]*schema.Resource, 0)
	for i, data := range d.Array() {
		name := fmt.Sprintf("ebs_block_device[%d]", i)
		resources = append(resources, newEbsBlockDevice(name, data, region))
	}
	return resources
}

func newEbsBlockDevice(name string, d gjson.Result, region string) *schema.Resource {
	volumeApiName := "gp2"
	if d.Get("volume_type").Exists() {
		volumeApiName = d.Get("volume_type").String()
	}

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("volume_size").Exists() {
		gbVal = decimal.NewFromFloat(d.Get("volume_size").Float())
	}

	iopsVal := decimal.Zero
	if d.Get("iops").Exists() {
		iopsVal = decimal.NewFromFloat(d.Get("iops").Float())
	}

	return &schema.Resource{
		Name:           name,
		CostComponents: ebsVolumeCostComponents(region, volumeApiName, gbVal, iopsVal),
	}
}
