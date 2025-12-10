package aws

import (
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type LaunchTemplate struct {
	// "required" args that can't really be missing.
	Address                          string
	Region                           string
	AMI                              string
	OnDemandBaseCount                int64
	OnDemandPercentageAboveBaseCount int64
	Tenancy                          string
	InstanceType                     string
	EBSOptimized                     bool
	EnableMonitoring                 bool
	CPUCredits                       string

	// "optional" args, that may be empty depending on the resource config
	ElasticInferenceAcceleratorType *string
	RootBlockDevice                 *EBSVolume
	EBSBlockDevices                 []*EBSVolume

	// "usage" args
	// These are populated from the Autoscaling Group/EKS Node Group resource
	InstanceCount                 *int64  `infracost_usage:"instances"`
	OperatingSystem               *string `infracost_usage:"operating_system"`
	ReservedInstanceType          *string `infracost_usage:"reserved_instance_type"`
	ReservedInstanceTerm          *string `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string `infracost_usage:"reserved_instance_payment_option"`
	MonthlyCPUCreditHours         *int64  `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                     *int64  `infracost_usage:"vcpu_count"`
}

var LaunchTemplateUsageSchema = InstanceUsageSchema

func (r *LaunchTemplate) CoreType() string {
	return "LaunchTemplate"
}

func (r *LaunchTemplate) UsageSchema() []*schema.UsageItem {
	return LaunchTemplateUsageSchema
}

func (r *LaunchTemplate) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LaunchTemplate) BuildResource() *schema.Resource {
	if strings.ToLower(r.Tenancy) == "host" {
		logging.Logger.Warn().Msgf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Templates", r.Address)
		return nil
	} else if strings.ToLower(r.Tenancy) == "dedicated" {
		r.Tenancy = "Dedicated"
	} else {
		r.Tenancy = "Shared"
	}

	costComponents := make([]*schema.CostComponent, 0)

	instance := &Instance{
		Region:                          r.Region,
		Tenancy:                         r.Tenancy,
		AMI:                             r.AMI,
		InstanceType:                    r.InstanceType,
		EBSOptimized:                    r.EBSOptimized,
		EnableMonitoring:                r.EnableMonitoring,
		CPUCredits:                      r.CPUCredits,
		ElasticInferenceAcceleratorType: r.ElasticInferenceAcceleratorType,
		OperatingSystem:                 r.OperatingSystem,
		RootBlockDevice:                 r.RootBlockDevice,
		EBSBlockDevices:                 r.EBSBlockDevices,
		ReservedInstanceType:            r.ReservedInstanceType,
		ReservedInstanceTerm:            r.ReservedInstanceTerm,
		ReservedInstancePaymentOption:   r.ReservedInstancePaymentOption,
		MonthlyCPUCreditHours:           r.MonthlyCPUCreditHours,
		VCPUCount:                       r.VCPUCount,
	}
	instanceResource := instance.BuildResource()

	// Skip the Instance usage cost component since we will prepend these later with the correct purchase options and counts
	for _, costComponent := range instanceResource.CostComponents {
		if !strings.HasPrefix(costComponent.Name, "Instance usage") {
			costComponents = append(costComponents, costComponent)
		}
	}

	res := &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
		SubResources:   instanceResource.SubResources,
		EstimateUsage:  instanceResource.EstimateUsage,
	}

	instanceCount := int64(1)
	if r.InstanceCount != nil {
		instanceCount = *r.InstanceCount
	}

	schema.MultiplyQuantities(res, decimal.NewFromInt(instanceCount))

	onDemandCount, spotCount := r.calculateOnDemandAndSpotInstanceCounts()

	if spotCount > 0 {
		instance.PurchaseOption = "spot"
		c := instance.computeCostComponent()
		c.MonthlyQuantity = decimalPtr(c.MonthlyQuantity.Mul(decimal.NewFromInt(spotCount)))
		res.CostComponents = append([]*schema.CostComponent{c}, res.CostComponents...)
	}

	if onDemandCount > 0 {
		instance.PurchaseOption = "on_demand"
		c := instance.computeCostComponent()
		c.MonthlyQuantity = decimalPtr(c.MonthlyQuantity.Mul(decimal.NewFromInt(onDemandCount)))
		res.CostComponents = append([]*schema.CostComponent{c}, res.CostComponents...)
	}

	return res
}

func (r *LaunchTemplate) calculateOnDemandAndSpotInstanceCounts() (int64, int64) {
	instanceCount := int64(1)
	if r.InstanceCount != nil {
		instanceCount = *r.InstanceCount
	}

	onDemandInstanceCount := r.OnDemandBaseCount
	remainingCount := instanceCount - onDemandInstanceCount
	percMultiplier := decimal.NewFromInt(r.OnDemandPercentageAboveBaseCount).Div(decimal.NewFromInt(100))
	onDemandInstanceCount += decimal.NewFromInt(remainingCount).Mul(percMultiplier).Ceil().IntPart()
	spotInstanceCount := instanceCount - onDemandInstanceCount

	return onDemandInstanceCount, spotInstanceCount
}
