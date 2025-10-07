package aws

import (
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type LaunchConfiguration struct {
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
	// These are populated from the Autoscaling Group resource
	InstanceCount                 *int64  `infracost_usage:"instances"`
	OperatingSystem               *string `infracost_usage:"operating_system"`
	ReservedInstanceType          *string `infracost_usage:"reserved_instance_type"`
	ReservedInstanceTerm          *string `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string `infracost_usage:"reserved_instance_payment_option"`
	MonthlyCPUCreditHours         *int64  `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                     *int64  `infracost_usage:"vcpu_count"`
}

var LaunchConfigurationUsageSchema = InstanceUsageSchema

func (r *LaunchConfiguration) CoreType() string {
	return "LaunchConfiguration"
}

func (r *LaunchConfiguration) UsageSchema() []*schema.UsageItem {
	return LaunchConfigurationUsageSchema
}

func (r *LaunchConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LaunchConfiguration) BuildResource() *schema.Resource {
	if strings.ToLower(r.Tenancy) == "host" {
		logging.Logger.Warn().Msgf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", r.Address)
		return nil
	} else if strings.ToLower(r.Tenancy) == "dedicated" {
		r.Tenancy = "Dedicated"
	} else {
		r.Tenancy = "Shared"
	}

	instance := &Instance{
		Region:                          r.Region,
		Tenancy:                         r.Tenancy,
		PurchaseOption:                  r.PurchaseOption,
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

	res := &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: instanceResource.CostComponents,
		SubResources:   instanceResource.SubResources,
		EstimateUsage:  instanceResource.EstimateUsage,
	}

	qty := int64(1)
	if r.InstanceCount != nil {
		qty = *r.InstanceCount
	}
	schema.MultiplyQuantities(res, decimal.NewFromInt(qty))

	return res
}
