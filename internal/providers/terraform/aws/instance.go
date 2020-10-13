package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var defaultEC2InstanceMetricCount = 7

func GetInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_instance",
		Notes: []string{
			"Costs associated with non-standard Linux AMIs, such as Windows and RHEL are not supported.",
			"EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.",
			"If a root volume is not specified then an 8Gi gp2 volume is assumed.",
		},
		RFunc: NewInstance,
	}
}

func NewInstance(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	tenancy := "Shared"
	if d.Get("tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS EC2 instances", d.Address)
		return nil
	} else if d.Get("tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	region := d.Get("region").String()
	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, newRootBlockDevice(d.Get("root_block_device.0"), region))
	subResources = append(subResources, newEbsBlockDevices(d.Get("ebs_block_device"), region)...)

	costComponents := []*schema.CostComponent{computeCostComponent(d, "on_demand", tenancy)}
	if d.Get("ebs_optimized").Bool() {
		costComponents = append(costComponents, ebsOptimizedCostComponent(d))
	}
	if d.Get("monitoring").Bool() {
		costComponents = append(costComponents, detailedMonitoringCostComponent(d))
	}
	c := cpuCreditsCostComponent(d)
	if c != nil {
		costComponents = append(costComponents, c)
	}

	return &schema.Resource{
		Name:           d.Address,
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}

func computeCostComponent(d *schema.ResourceData, purchaseOption string, tenancy string) *schema.CostComponent {
	region := d.Get("region").String()
	instanceType := d.Get("instance_type").String()

	purchaseOptionLabel := map[string]string{
		"on_demand": "on-demand",
		"spot":      "spot",
	}[purchaseOption]

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Linux/UNIX usage (%s, %s)", purchaseOptionLabel, instanceType),
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
	}
}

func ebsOptimizedCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	instanceType := d.Get("instance_type").String()

	return &schema.CostComponent{
		Name:                 "EBS-optimized usage",
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
	}
}

func detailedMonitoringCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()

	return &schema.CostComponent{
		Name:                 "EC2 detailed monitoring",
		Unit:                 "metrics",
		MonthlyQuantity:      decimalPtr(decimal.NewFromInt(int64(defaultEC2InstanceMetricCount))),
		IgnoreIfMissingPrice: true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Metric"),
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func cpuCreditsCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	instanceType := d.Get("instance_type").String()

	cpuCredits := d.Get("credit_specification.0.cpu_credits").String()
	if cpuCredits == "" && (strings.HasPrefix(instanceType, "t3.") || strings.HasPrefix(instanceType, "t4g.")) {
		cpuCredits = "unlimited"
	}

	if cpuCredits != "unlimited" {
		return nil
	}

	prefix := strings.SplitN(instanceType, ".", 2)[0]

	return &schema.CostComponent{
		Name:           "CPU credits",
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
	}
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
	volumeAPIName := "gp2"
	if d.Get("volume_type").Exists() {
		volumeAPIName = d.Get("volume_type").String()
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
		CostComponents: ebsVolumeCostComponents(region, volumeAPIName, gbVal, iopsVal),
	}
}
