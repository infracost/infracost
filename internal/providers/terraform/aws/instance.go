package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
			"Costs associated with marketplace AMIs are not supported.",
			"For non-standard Linux AMIs such as Windows and RHEL, the operating system should be specified in usage file.",
			"EC2 detailed monitoring assumes the standard 7 metrics and the lowest tier of prices for CloudWatch.",
			"If a root volume is not specified then an 8Gi gp2 volume is assumed.",
		},
		RFunc: NewInstance,
	}
}

func NewInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	tenancy := "Shared"
	if d.Get("tenancy").String() == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS EC2 instances", d.Address)
		return nil
	} else if d.Get("tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	instanceType := d.Get("instance_type").String()

	region := d.Get("region").String()
	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, newRootBlockDevice(d.Get("root_block_device.0"), region))
	subResources = append(subResources, newEbsBlockDevices(d.Get("ebs_block_device"), region)...)

	costComponents := []*schema.CostComponent{computeCostComponent(d, u, "on_demand", instanceType, tenancy, 1)}
	if d.Get("ebs_optimized").Bool() {
		costComponents = append(costComponents, ebsOptimizedCostComponent(d))
	}
	if d.Get("monitoring").Bool() {
		costComponents = append(costComponents, detailedMonitoringCostComponent(d))
	}

	if isInstanceBurstable(d.Get("instance_type").String(), []string{"t2.", "t3.", "t4."}) {
		c := newCPUCredit(d, u)
		if c != nil {
			costComponents = append(costComponents, c)
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		SubResources:   subResources,
		CostComponents: costComponents,
	}
}

func computeCostComponent(d *schema.ResourceData, u *schema.UsageData, purchaseOption, instanceType, tenancy string, desiredSize int64) *schema.CostComponent {
	region := d.Get("region").String()

	purchaseOptionLabel := map[string]string{
		"on_demand": "on-demand",
		"spot":      "spot",
	}[purchaseOption]

	osLabel := "Linux/UNIX"
	operatingSystem := "Linux"
	var usageOperation string

	// Allow the operating system to be specified in the usage data until we can support it from the AMI directly.
	if u != nil && u.Get("operating_system").Exists() {
		os := strings.ToLower(u.Get("operating_system").String())
		switch os {
		case "windows":
			osLabel = "Windows"
			operatingSystem = "Windows"
		case "rhel":
			osLabel = "RHEL"
			operatingSystem = "RHEL"
		case "suse":
			osLabel = "SUSE"
			operatingSystem = "SUSE"
		default:
			if os != "linux" {
				log.Warnf("Unrecognized operating system %s, defaulting to Linux/UNIX", os)
			}
		}
	} else {
		ami := d.Get("image_id").String()
		sess, err := session.NewSession(&aws.Config{
			Region: &region,
		})
		if err == nil {
			svc := ec2.New(sess)
			input := &ec2.DescribeImagesInput{
				ImageIds: []*string{
					aws.String(ami),
				},
			}
			result, err := svc.DescribeImages(input)
			if err == nil {
				if len(result.Images) > 0 {
					usageOperation = *result.Images[0].UsageOperation
				}
			}
		}
	}

	var reservedType, reservedTerm, reservedPaymentOption string
	if u != nil && u.Get("reserved_instance_type").Type != gjson.Null &&
		u.Get("reserved_instance_term").Type != gjson.Null &&
		u.Get("reserved_instance_payment_option").Type != gjson.Null {

		reservedType = u.Get("reserved_instance_type").String()
		reservedTerm = u.Get("reserved_instance_term").String()
		reservedPaymentOption = u.Get("reserved_instance_payment_option").String()

		valid, err := validateReserveInstanceParams(reservedType, reservedTerm, reservedPaymentOption)
		if err != "" {
			log.Warnf(err)
		}
		if valid {
			purchaseOptionLabel = "reserved"
			return reservedInstanceCostComponent(region, osLabel, purchaseOptionLabel, reservedType, reservedTerm, reservedPaymentOption, tenancy, instanceType, operatingSystem, 1)
		}
	}

	costComponent := &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, purchaseOptionLabel, instanceType),
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
				{Key: "tenancy", Value: strPtr(tenancy)},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "licenseModel", Value: strPtr("No License required")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: &purchaseOption,
		},
	}

	if usageOperation != "" {
		costComponent.ProductFilter.AttributeFilters = append(costComponent.ProductFilter.AttributeFilters,
			&schema.AttributeFilter{Key: "operation", Value: strPtr(usageOperation)})
	} else {
		costComponent.ProductFilter.AttributeFilters = append(costComponent.ProductFilter.AttributeFilters,
			&schema.AttributeFilter{Key: "operatingSystem", Value: strPtr(operatingSystem)})
	}

	return costComponent
}

func validateReserveInstanceParams(typeName, term, option string) (bool, string) {
	validTypes := []string{"convertible", "standard"}
	if !stringInSlice(validTypes, typeName) {
		return false, fmt.Sprintf("Invalid reserved_instance_type, ignoring reserved options. Expected: convertible, standard. Got: %s", typeName)
	}

	validTerms := []string{"1_year", "3_year"}
	if !stringInSlice(validTerms, term) {
		return false, fmt.Sprintf("Invalid reserved_instance_term, ignoring reserved options. Expected: 1_year, 3_year. Got: %s", term)
	}

	validOptions := []string{"no_upfront", "partial_upfront", "all_upfront"}
	if !stringInSlice(validOptions, option) {
		return false, fmt.Sprintf("Invalid reserved_instance_payment_option, ignoring reserved options. Expected: no_upfront, partial_upfront, all_upfront. Got: %s", option)
	}

	return true, ""
}

func reservedInstanceCostComponent(region, osLabel, purchaseOptionLabel, reservedType, reservedTerm, reservedPaymentOption, tenancy, instanceType, operatingSystem string, count int64) *schema.CostComponent {
	reservedTermName := map[string]string{
		"1_year": "1yr",
		"3_year": "3yr",
	}[reservedTerm]

	reservedPaymentOptionName := map[string]string{
		"no_upfront":      "No Upfront",
		"partial_upfront": "Partial Upfront",
		"all_upfront":     "All Upfront",
	}[reservedPaymentOption]

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, purchaseOptionLabel, instanceType),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(count)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(instanceType)},
				{Key: "tenancy", Value: strPtr(tenancy)},
				{Key: "operatingSystem", Value: strPtr(operatingSystem)},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount:   strPtr("0"),
			TermOfferingClass:  &reservedType,
			TermLength:         &reservedTermName,
			TermPurchaseOption: &reservedPaymentOptionName,
		},
	}
}

func ebsOptimizedCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	instanceType := d.Get("instance_type").String()

	return &schema.CostComponent{
		Name:                 "EBS-optimized usage",
		Unit:                 "hours",
		UnitMultiplier:       1,
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
		UnitMultiplier:       1,
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

func cpuCreditsCostComponent(region string, vCPUCount decimal.Decimal, prefix string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "CPU credits",
		Unit:            "vCPU-hours",
		UnitMultiplier:  1,
		MonthlyQuantity: &vCPUCount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("CPU Credits"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operatingSystem", Value: strPtr("Linux")},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/CPUCredits:%s$/", prefix))},
			},
		},
	}
}

func isInstanceBurstable(instanceType string, burstableInstanceTypes []string) bool {
	for _, instance := range burstableInstanceTypes {
		if strings.HasPrefix(instanceType, instance) {
			return true
		}
	}
	return false
}

func newCPUCredit(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
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

	instanceCPUCreditHours := decimal.Zero
	if u != nil && u.Get("monthly_cpu_credit_hrs").Exists() {
		instanceCPUCreditHours = decimal.NewFromInt(u.Get("monthly_cpu_credit_hrs").Int())
	}

	instanceVCPUCount := decimal.Zero
	if u != nil && u.Get("vcpu_count").Exists() {
		instanceVCPUCount = decimal.NewFromInt(u.Get("vcpu_count").Int())
	}

	cpuCreditQuantity := instanceVCPUCount.Mul(instanceCPUCreditHours)

	return cpuCreditsCostComponent(region, cpuCreditQuantity, prefix)
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

	var unknown *decimal.Decimal

	return &schema.Resource{
		Name:           name,
		CostComponents: ebsVolumeCostComponents(region, volumeAPIName, unknown, gbVal, iopsVal, unknown),
	}
}
