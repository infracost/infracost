package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var defaultEC2InstanceMetricCount = 7
var burstableInstanceTypePrefixes = []string{"t2", "t3", "t3a", "t4g"}

type Instance struct {
	// "required" args that can't really be missing.
	Address          string
	Region           string
	Tenancy          string
	PurchaseOption   string
	AMI              string
	InstanceType     string
	EBSOptimized     bool
	EnableMonitoring bool
	CPUCredits       string

	// "optional" args, that may be empty depending on the resource config
	ElasticInferenceAcceleratorType *string
	RootBlockDevice                 *EBSVolume
	EBSBlockDevices                 []*EBSVolume

	// "usage" args
	OperatingSystem               *string `infracost_usage:"operating_system"`
	ReservedInstanceType          *string `infracost_usage:"reserved_instance_type"`
	ReservedInstanceTerm          *string `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string `infracost_usage:"reserved_instance_payment_option"`
	MonthlyCPUCreditHours         *int64  `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                     *int64  `infracost_usage:"vcpu_count"`
}

var InstanceUsageSchema = []*schema.UsageItem{
	{Key: "operating_system", DefaultValue: "linux", ValueType: schema.String},
	{Key: "reserved_instance_type", DefaultValue: "", ValueType: schema.String},
	{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
	{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
	{Key: "monthly_cpu_credit_hrs", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "vcpu_count", DefaultValue: 0, ValueType: schema.Int64},
}

func (a *Instance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *Instance) BuildResource() *schema.Resource {
	if strings.ToLower(a.Tenancy) == "host" {
		log.Warnf("Skipping resource %s. Infracost currently does not support host tenancy for AWS EC2 instances", a.Address)
		return nil
	} else if strings.ToLower(a.Tenancy) == "dedicated" {
		a.Tenancy = "Dedicated"
	} else {
		a.Tenancy = "Shared"
	}

	if a.CPUCredits == "" && (strings.HasPrefix(a.InstanceType, "t3.") || strings.HasPrefix(a.InstanceType, "t4g.")) {
		a.CPUCredits = "unlimited"
	}

	if a.OperatingSystem == nil {
		a.OperatingSystem = strPtr("linux")
	}

	if a.PurchaseOption == "" {
		a.PurchaseOption = "on_demand"
	}

	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	if a.RootBlockDevice != nil {
		subResources = append(subResources, a.RootBlockDevice.BuildResource())
	}

	for _, ebs := range a.EBSBlockDevices {
		subResources = append(subResources, ebs.BuildResource())
	}

	costComponents = append(costComponents, a.computeCostComponent())

	if a.EBSOptimized {
		costComponents = append(costComponents, a.ebsOptimizedCostComponent())
	}

	if a.EnableMonitoring {
		costComponents = append(costComponents, a.detailedMonitoringCostComponent())
	}

	if a.ElasticInferenceAcceleratorType != nil {
		costComponents = append(costComponents, a.elasticInferenceAcceleratorCostComponent())
	}

	if a.CPUCredits == "unlimited" {
		if instanceFamily := getBurstableInstanceFamily(burstableInstanceTypePrefixes, a.InstanceType); instanceFamily != "" {
			costComponents = append(costComponents, a.cpuCreditCostComponent(instanceFamily))
		}
	}

	estimate := func(ctx context.Context, values map[string]interface{}) error {
		if a.AMI != "" {
			platform, err := aws.EC2DescribeOS(ctx, a.Region, a.AMI)
			if err != nil {
				return err
			}
			if platform != "" {
				values["operating_system"] = platform
			}
		}
		return nil
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    InstanceUsageSchema,
		CostComponents: costComponents,
		SubResources:   subResources,
		EstimateUsage:  estimate,
	}
}

func (a *Instance) computeCostComponent() *schema.CostComponent {
	purchaseOptionLabel := map[string]string{
		"on_demand": "on-demand",
		"spot":      "spot",
	}[a.PurchaseOption]

	osLabel := "Linux/UNIX"
	osFilterVal := "Linux"

	switch strVal(a.OperatingSystem) {
	case "windows":
		osLabel = "Windows"
		osFilterVal = "Windows"
	case "rhel":
		osLabel = "RHEL"
		osFilterVal = "RHEL"
	case "suse":
		osLabel = "SUSE"
		osFilterVal = "SUSE"
	default:
		if strVal(a.OperatingSystem) != "linux" {
			log.Warnf("Unrecognized operating system %s, defaulting to Linux/UNIX", strVal(a.OperatingSystem))
		}
	}

	if a.ReservedInstanceType != nil {
		valid, err := a.validateReserveInstanceParams()
		if err != "" {
			log.Warnf(err)
		}
		if valid {
			purchaseOptionLabel = "reserved"
			return a.reservedInstanceCostComponent(osLabel, osFilterVal, purchaseOptionLabel)
		}
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, purchaseOptionLabel, a.InstanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(a.InstanceType)},
				{Key: "tenancy", Value: strPtr(a.Tenancy)},
				{Key: "operatingSystem", Value: strPtr(osFilterVal)},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "licenseModel", Value: strPtr("No License required")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(a.PurchaseOption),
		},
	}
}

func (a *Instance) validateReserveInstanceParams() (bool, string) {
	validTypes := []string{"convertible", "standard"}
	if !stringInSlice(validTypes, strVal(a.ReservedInstanceType)) {
		return false, fmt.Sprintf("Invalid reserved_instance_type, ignoring reserved options. Expected: convertible, standard. Got: %s", strVal(a.ReservedInstanceType))
	}

	validTerms := []string{"1_year", "3_year"}
	if !stringInSlice(validTerms, strVal(a.ReservedInstanceTerm)) {
		return false, fmt.Sprintf("Invalid reserved_instance_term, ignoring reserved options. Expected: 1_year, 3_year. Got: %s", strVal(a.ReservedInstanceTerm))
	}

	validOptions := []string{"no_upfront", "partial_upfront", "all_upfront"}
	if !stringInSlice(validOptions, strVal(a.ReservedInstancePaymentOption)) {
		return false, fmt.Sprintf("Invalid reserved_instance_payment_option, ignoring reserved options. Expected: no_upfront, partial_upfront, all_upfront. Got: %s", strVal(a.ReservedInstancePaymentOption))
	}

	return true, ""
}

func (a *Instance) reservedInstanceCostComponent(osLabel, osFilterVal, purchaseOptionLabel string) *schema.CostComponent {
	reservedTermName := map[string]string{
		"1_year": "1yr",
		"3_year": "3yr",
	}[strVal(a.ReservedInstanceTerm)]

	reservedPaymentOptionName := map[string]string{
		"no_upfront":      "No Upfront",
		"partial_upfront": "Partial Upfront",
		"all_upfront":     "All Upfront",
	}[strVal(a.ReservedInstancePaymentOption)]

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, purchaseOptionLabel, a.InstanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(a.InstanceType)},
				{Key: "tenancy", Value: strPtr(a.Tenancy)},
				{Key: "operatingSystem", Value: strPtr(osFilterVal)},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount:   strPtr("0"),
			TermOfferingClass:  a.ReservedInstanceType,
			TermLength:         strPtr(reservedTermName),
			TermPurchaseOption: strPtr(reservedPaymentOptionName),
		},
	}
}

func (a *Instance) ebsOptimizedCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:                 "EBS-optimized usage",
		Unit:                 "hours",
		UnitMultiplier:       decimal.NewFromInt(1),
		HourlyQuantity:       decimalPtr(decimal.NewFromInt(1)),
		IgnoreIfMissingPrice: true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(a.InstanceType)},
				{Key: "usagetype", ValueRegex: strPtr("/EBSOptimized/")},
			},
		},
	}
}

func (a *Instance) detailedMonitoringCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:                 "EC2 detailed monitoring",
		Unit:                 "metrics",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      decimalPtr(decimal.NewFromInt(int64(defaultEC2InstanceMetricCount))),
		IgnoreIfMissingPrice: true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Metric"),
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func (a *Instance) elasticInferenceAcceleratorCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Inference accelerator (%s)", strVal(a.ElasticInferenceAcceleratorType)),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEI"),
			ProductFamily: strPtr("Elastic Inference"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", strVal(a.ElasticInferenceAcceleratorType)))},
			},
		},
	}
}

func (a *Instance) cpuCreditCostComponent(instanceFamily string) *schema.CostComponent {
	qty := decimal.Zero
	if a.MonthlyCPUCreditHours != nil && a.VCPUCount != nil {
		qty = decimal.NewFromInt(*a.MonthlyCPUCreditHours).Mul(decimal.NewFromInt(*a.VCPUCount))
	}

	return &schema.CostComponent{
		Name:            "CPU credits",
		Unit:            "vCPU-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(qty),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("CPU Credits"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operatingSystem", Value: strPtr("Linux")},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/CPUCredits:%s$/", instanceFamily))},
			},
		},
	}
}
