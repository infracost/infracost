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

func (a *LaunchTemplate) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *LaunchTemplate) BuildResource() *schema.Resource {
	if strings.ToLower(a.Tenancy) == "host" {
		logging.Logger.Warn().Msgf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Templates", a.Address)
		return nil
	} else if strings.ToLower(a.Tenancy) == "dedicated" {
		a.Tenancy = "Dedicated"
	} else {
		a.Tenancy = "Shared"
	}

	costComponents := make([]*schema.CostComponent, 0)

	instance := &Instance{
		Region:                          a.Region,
		Tenancy:                         a.Tenancy,
		AMI:                             a.AMI,
		InstanceType:                    a.InstanceType,
		EBSOptimized:                    a.EBSOptimized,
		EnableMonitoring:                a.EnableMonitoring,
		CPUCredits:                      a.CPUCredits,
		ElasticInferenceAcceleratorType: a.ElasticInferenceAcceleratorType,
		OperatingSystem:                 a.OperatingSystem,
		RootBlockDevice:                 a.RootBlockDevice,
		EBSBlockDevices:                 a.EBSBlockDevices,
		ReservedInstanceType:            a.ReservedInstanceType,
		ReservedInstanceTerm:            a.ReservedInstanceTerm,
		ReservedInstancePaymentOption:   a.ReservedInstancePaymentOption,
		MonthlyCPUCreditHours:           a.MonthlyCPUCreditHours,
		VCPUCount:                       a.VCPUCount,
	}
	instanceResource := instance.BuildResource()

	// Skip the Instance usage cost component since we will prepend these later with the correct purchase options and counts
	for _, costComponent := range instanceResource.CostComponents {
		if !strings.HasPrefix(costComponent.Name, "Instance usage") {
			costComponents = append(costComponents, costComponent)
		}
	}

	r := &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		CostComponents: costComponents,
		SubResources:   instanceResource.SubResources,
		EstimateUsage:  instanceResource.EstimateUsage,
	}

	instanceCount := int64(1)
	if a.InstanceCount != nil {
		instanceCount = *a.InstanceCount
	}

	schema.MultiplyQuantities(r, decimal.NewFromInt(instanceCount))

	onDemandCount, spotCount := a.calculateOnDemandAndSpotInstanceCounts()

	if spotCount > 0 {
		instance.PurchaseOption = "spot"
		c := instance.computeCostComponent()
		c.MonthlyQuantity = decimalPtr(c.MonthlyQuantity.Mul(decimal.NewFromInt(spotCount)))
		r.CostComponents = append([]*schema.CostComponent{c}, r.CostComponents...)
	}

	if onDemandCount > 0 {
		instance.PurchaseOption = "on_demand"
		c := instance.computeCostComponent()
		c.MonthlyQuantity = decimalPtr(c.MonthlyQuantity.Mul(decimal.NewFromInt(onDemandCount)))
		r.CostComponents = append([]*schema.CostComponent{c}, r.CostComponents...)
	}

	return r
}

func (a *LaunchTemplate) calculateOnDemandAndSpotInstanceCounts() (int64, int64) {
	instanceCount := int64(1)
	if a.InstanceCount != nil {
		instanceCount = *a.InstanceCount
	}

	onDemandInstanceCount := a.OnDemandBaseCount
	remainingCount := instanceCount - onDemandInstanceCount
	percMultiplier := decimal.NewFromInt(a.OnDemandPercentageAboveBaseCount).Div(decimal.NewFromInt(100))
	onDemandInstanceCount += decimal.NewFromInt(remainingCount).Mul(percMultiplier).Ceil().IntPart()
	spotInstanceCount := instanceCount - onDemandInstanceCount

	return onDemandInstanceCount, spotInstanceCount
}
