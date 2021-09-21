package aws

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

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
			if u != nil && u.Get("monthly_cpu_credit_hrs").Exists() {
				instanceCPUCreditHours = decimal.NewFromInt(u.Get("monthly_cpu_credit_hrs").Int())
			}

			instanceVCPUCount := decimal.Zero
			if u != nil && u.Get("vcpu_count").Exists() {
				instanceVCPUCount = decimal.NewFromInt(u.Get("vcpu_count").Int())
			}

			cpuCreditQuantity = instanceVCPUCount.Mul(instanceCPUCreditHours).Mul(decimal.NewFromInt(desiredSize))
			instancePrefix := strings.SplitN(instanceType, ".", 2)[0]
			costComponents = append(costComponents, cpuCreditsCostComponent(region, cpuCreditQuantity, instancePrefix))
		}

		costComponents = append(costComponents, newEksRootBlockDevice(d))
	}

	if len(launchTemplateRef) > 0 {
		spotCount := int64(0)
		onDemandCount := desiredSize

		if strings.ToLower(launchTemplateRef[0].Get("instance_market_options.0.market_type").String()) == "spot" {
			onDemandCount = int64(0)
			spotCount = desiredSize
		}

		if launchTemplateRef[0].Get("instance_type").Type == gjson.Null {
			launchTemplateRef[0].Set("instance_type", d.Get("instance_types").Array()[0].String())
		}

		lt := newLaunchTemplate(launchTemplateRef[0], u, region, onDemandCount, spotCount)
		ltResource := lt.BuildResource()

		// AutoscalingGroup should show as not supported LaunchTemplate is not supported
		if ltResource == nil {
			return nil
		}
		subResources = append(subResources, ltResource)
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

// TODO: these are copied from the old instance resource and can be removed when this resource is moved over to the new structure

func computeCostComponent(d *schema.ResourceData, u *schema.UsageData, purchaseOption, instanceType, tenancy string, desiredSize int64) *schema.CostComponent {
	region := d.Get("region").String()

	purchaseOptionLabel := map[string]string{
		"on_demand": "on-demand",
		"spot":      "spot",
	}[purchaseOption]

	osLabel := "Linux/UNIX"
	operatingSystem := "Linux"

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

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, purchaseOptionLabel, instanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(desiredSize)),
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
				{Key: "licenseModel", Value: strPtr("No License required")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: &purchaseOption,
		},
	}
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
		UnitMultiplier: decimal.NewFromInt(1),
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

func cpuCreditsCostComponent(region string, vCPUCount decimal.Decimal, prefix string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "CPU credits",
		Unit:            "vCPU-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
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
