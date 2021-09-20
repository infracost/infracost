package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type LaunchConfiguration struct {
	// "required" args that can't really be missing.
	Address          string
	Region           string
	Tenancy          string
	PurchaseOption   string
	InstanceType     string
	EBSOptimized     bool
	EnableMonitoring bool
	CPUCredits       string

	// "optional" args, that may be empty depending on the resource config
	RootBlockDevice *EBSVolume
	EBSBlockDevices []*EBSVolume

	// "usage" args
	OperatingSystem               *string `infracost_usage:"operating_system"`
	ReservedInstanceType          *string `infracost_usage:"reserved_instance_type"`
	ReservedInstanceTerm          *string `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string `infracost_usage:"reserved_instance_payment_option"`
	MonthlyCPUCreditHours         *int64  `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                     *int64  `infracost_usage:"vcpu_count"`
}

var LaunchConfigurationUsageSchema = InstanceUsageSchema

func (a *LaunchConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *LaunchConfiguration) BuildResource() *schema.Resource {
	instance := &Instance{
		Region:                        a.Region,
		Tenancy:                       a.Tenancy,
		PurchaseOption:                a.PurchaseOption,
		InstanceType:                  a.InstanceType,
		EBSOptimized:                  a.EBSOptimized,
		EnableMonitoring:              a.EnableMonitoring,
		CPUCredits:                    a.CPUCredits,
		OperatingSystem:               a.OperatingSystem,
		RootBlockDevice:               a.RootBlockDevice,
		EBSBlockDevices:               a.EBSBlockDevices,
		ReservedInstanceType:          a.ReservedInstanceType,
		ReservedInstanceTerm:          a.ReservedInstanceTerm,
		ReservedInstancePaymentOption: a.ReservedInstancePaymentOption,
		MonthlyCPUCreditHours:         a.MonthlyCPUCreditHours,
		VCPUCount:                     a.VCPUCount,
	}
	instanceResource := instance.BuildResource()

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    LaunchConfigurationUsageSchema,
		CostComponents: instanceResource.CostComponents,
		SubResources:   instanceResource.SubResources,
	}
}
