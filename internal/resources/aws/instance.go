package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
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
	HasHost          bool

	// "optional" args, that may be empty depending on the resource config
	ElasticInferenceAcceleratorType *string
	RootBlockDevice                 *EBSVolume
	EBSBlockDevices                 []*EBSVolume

	// "usage" args
	OperatingSystem               *string  `infracost_usage:"operating_system"`
	ReservedInstanceType          *string  `infracost_usage:"reserved_instance_type"`
	ReservedInstanceTerm          *string  `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string  `infracost_usage:"reserved_instance_payment_option"`
	MonthlyCPUCreditHours         *int64   `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                     *int64   `infracost_usage:"vcpu_count"`
	MonthlyHours                  *float64 `infracost_usage:"monthly_hrs"`
}

func (a *Instance) CoreType() string {
	return "Instance"
}

func (a *Instance) UsageSchema() []*schema.UsageItem {
	return InstanceUsageSchema
}

var InstanceUsageSchema = []*schema.UsageItem{
	{Key: "operating_system", DefaultValue: "linux", ValueType: schema.String},
	{Key: "reserved_instance_type", DefaultValue: "", ValueType: schema.String},
	{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
	{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
	{Key: "monthly_cpu_credit_hrs", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "vcpu_count", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_hrs", DefaultValue: 730, ValueType: schema.Float64},
}

func (a *Instance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *Instance) BuildResource() *schema.Resource {
	if strings.ToLower(a.Tenancy) == "host" {
		if a.HasHost {
			a.Tenancy = "Host"
		} else {
			logging.Logger.Warn().Msgf("Skipping resource %s. Infracost currently does not support host tenancy for AWS EC2 instances without Host ID set up", a.Address)
			return nil
		}
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

	if !a.HasHost {
		costComponents = append(costComponents, a.computeCostComponent())
	}

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

	estimate := func(ctx context.Context, values map[string]any) error {
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
		UsageSchema:    a.UsageSchema(),
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
			logging.Logger.Warn().Msgf("Unrecognized operating system %s, defaulting to Linux/UNIX", strVal(a.OperatingSystem))
		}
	}

	priceFilter := &schema.PriceFilter{
		PurchaseOption: strPtr(a.PurchaseOption),
	}

	if a.ReservedInstanceType != nil {
		resolver := &ec2ReservationResolver{
			term:              strVal(a.ReservedInstanceTerm),
			paymentOption:     strVal(a.ReservedInstancePaymentOption),
			termOfferingClass: strVal(a.ReservedInstanceType),
		}
		reservedFilter, err := resolver.PriceFilter()
		if err != nil {
			logging.Logger.Warn().Msg(err.Error())
		} else {
			priceFilter = reservedFilter
		}
		purchaseOptionLabel = "reserved"
	}

	qty := decimal.NewFromFloat(730)
	if a.MonthlyHours != nil {
		qty = decimal.NewFromFloat(*a.MonthlyHours)
	}

	// metal instances have a different ProductFamily in AWS pricing data
	productFamily := "Compute Instance"
	if strings.Contains(strings.ToLower(a.InstanceType), "metal") {
		productFamily = "Compute Instance (bare metal)"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance usage (%s, %s, %s)", osLabel, purchaseOptionLabel, a.InstanceType),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(qty),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr(productFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(a.InstanceType)},
				{Key: "tenancy", Value: strPtr(a.Tenancy)},
				{Key: "operatingSystem", Value: strPtr(osFilterVal)},
				{Key: "preInstalledSw", Value: strPtr("NA")},
				{Key: "licenseModel", Value: strPtr("No License required")},
				{Key: "capacitystatus", Value: strPtr("Used")},
			},
		},
		PriceFilter: priceFilter,
	}
}

func (a *Instance) ebsOptimizedCostComponent() *schema.CostComponent {
	/**
	 * EBS Optimized instances are billed hourly whenever the attached instance is live.
	 *
	 * From the EBS-opimized instance docs:
	 *    > For Current Generation Instance types, EBS-optimization is enabled by default
	 *    > at no additional cost. For Previous Generation Instances types, EBS-optimization
	 *    > prices are on the Previous Generation Pricing Page.
	 *    >
	 *    > The hourly price for EBS-optimized instances is in addition to the hourly usage fee
	 *    > for supported instance types.
	 */
	qty := decimal.NewFromFloat(730)
	if a.MonthlyHours != nil {
		qty = decimal.NewFromFloat(*a.MonthlyHours)
	}

	return &schema.CostComponent{
		Name:                 "EBS-optimized usage",
		Unit:                 "hours",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      decimalPtr(qty),
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
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", Value: strPtr("Metric")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func (a *Instance) elasticInferenceAcceleratorCostComponent() *schema.CostComponent {
	/**
	 * Elastic inference accelerators are billed hourly whenever the attached instance is live.
	 *
	 * From the elastic inference accelerator  docs:
	 *    > With Amazon Elastic Inference, you pay only for the accelerator hours you use.
	 *    > There are no upfront costs or minimum fees.
	 */
	qty := decimal.NewFromFloat(730)
	if a.MonthlyHours != nil {
		qty = decimal.NewFromFloat(*a.MonthlyHours)
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Inference accelerator (%s)", strVal(a.ElasticInferenceAcceleratorType)),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(qty),
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

type ec2ReservationResolver struct {
	term              string
	paymentOption     string
	termOfferingClass string
}

// PriceFilter implementation for ec2ReservationResolver
// Allowed values for ReservedInstanceTerm: ["1_year", "3_year"]
// Allowed values for ReservedInstancePaymentOption: ["all_upfront", "partial_upfront", "no_upfront"]
// Allowed values for ReservedTermOfferingClass: ["standard", "convertible"]
func (r ec2ReservationResolver) PriceFilter() (*schema.PriceFilter, error) {
	termLength := reservedTermsMapping[r.term]
	purchaseOption := reservedPaymentOptionMapping[r.paymentOption]
	validTypes := []string{"convertible", "standard"}
	if !stringInSlice(validTypes, r.termOfferingClass) {
		return nil, fmt.Errorf("Invalid reserved_instance_type, ignoring reserved options. Expected: convertible, standard. Got: %s", r.termOfferingClass)
	}
	validTerms := sliceOfKeysFromMap(reservedTermsMapping)
	if !stringInSlice(validTerms, r.term) {
		return nil, fmt.Errorf("Invalid reserved_instance_term, ignoring reserved options. Expected: %s. Got: %s", strings.Join(validTerms, ", "), r.term)
	}
	validOptions := sliceOfKeysFromMap(reservedPaymentOptionMapping)
	if !stringInSlice(validOptions, r.paymentOption) {
		return nil, fmt.Errorf("Invalid reserved_instance_payment_option, ignoring reserved options. Expected: %s. Got: %s", strings.Join(validOptions, ", "), r.paymentOption)
	}
	return &schema.PriceFilter{
		StartUsageAmount:   strPtr("0"),
		TermOfferingClass:  strPtr(r.termOfferingClass),
		TermLength:         strPtr(termLength),
		TermPurchaseOption: strPtr(purchaseOption),
	}, nil
}
