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

func (a *LaunchConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *LaunchConfiguration) BuildResource() *schema.Resource {
	if strings.ToLower(a.Tenancy) == "host" {
		logging.Logger.Warn().Msgf("Skipping resource %s. Infracost currently does not support host tenancy for AWS Launch Configurations", a.Address)
		return nil
	} else if strings.ToLower(a.Tenancy) == "dedicated" {
		a.Tenancy = "Dedicated"
	} else {
		a.Tenancy = "Shared"
	}

	instance := &Instance{
		Region:                          a.Region,
		Tenancy:                         a.Tenancy,
		PurchaseOption:                  a.PurchaseOption,
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

	r := &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		CostComponents: instanceResource.CostComponents,
		SubResources:   instanceResource.SubResources,
		EstimateUsage:  instanceResource.EstimateUsage,
	}

	qty := int64(1)
	if a.InstanceCount != nil {
		qty = *a.InstanceCount
	}
	schema.MultiplyQuantities(r, decimal.NewFromInt(qty))

	return r
}
