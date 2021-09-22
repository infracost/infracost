package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type EKSNodeGroup struct {
	// "required" args that can't really be missing.
	Address string
	Region  string

	InstanceType   string
	PurchaseOption string
	DiskSize       int64

	// "optional" args, that may be empty depending on the resource config
	RootBlockDevice *EBSVolume
	LaunchTemplate  *LaunchTemplate

	// "usage" args
	InstanceCount                 *int64  `infracost_usage:"instances"`
	OperatingSystem               *string `infracost_usage:"operating_system"`
	ReservedInstanceType          *string `infracost_usage:"reserved_instance_type"`
	ReservedInstanceTerm          *string `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string `infracost_usage:"reserved_instance_payment_option"`
	MonthlyCPUCreditHours         *int64  `infracost_usage:"monthly_cpu_credit_hrs"`
	VCPUCount                     *int64  `infracost_usage:"vcpu_count"`
}

var EKSNodeGroupUsageSchema = append([]*schema.UsageSchemaItem{
	{Key: "instances", DefaultValue: 0, ValueType: schema.Int64},
}, InstanceUsageSchema...)

func (a *EKSNodeGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)

	// The usage keys for Launch Template are specified on the EKS Node Groupresource
	if a.LaunchTemplate != nil {
		resources.PopulateArgsWithUsage(a.LaunchTemplate, u)
	}
}

func (a *EKSNodeGroup) BuildResource() *schema.Resource {
	r := &schema.Resource{
		Name:        a.Address,
		UsageSchema: EKSNodeGroupUsageSchema,
	}

	if a.LaunchTemplate != nil {
		lt := a.LaunchTemplate.BuildResource()
		// If the Launch Template returns nil it is not supported so the Autoscaling Group should also return nil
		if lt == nil {
			return nil
		}
		r.SubResources = append(r.SubResources, lt)
	} else {
		instance := &Instance{
			Region:                        a.Region,
			Tenancy:                       "Shared",
			InstanceType:                  a.InstanceType,
			PurchaseOption:                a.PurchaseOption,
			OperatingSystem:               a.OperatingSystem,
			ReservedInstanceType:          a.ReservedInstanceType,
			ReservedInstanceTerm:          a.ReservedInstanceTerm,
			ReservedInstancePaymentOption: a.ReservedInstancePaymentOption,
			MonthlyCPUCreditHours:         a.MonthlyCPUCreditHours,
			VCPUCount:                     a.VCPUCount,
		}

		instance.RootBlockDevice = &EBSVolume{
			Address: "root_block_device",
			Region:  a.Region,
			Type:    "gp2",
			Size:    intPtr(a.DiskSize),
		}

		instanceResource := instance.BuildResource()
		r.CostComponents = append(r.CostComponents, instanceResource.CostComponents...)

		// For EKS Node Groups we show the root block device cost component into the top level of the resource
		for _, subResource := range instanceResource.SubResources {
			if subResource.Name == "root_block_device" {
				r.CostComponents = append(r.CostComponents, subResource.CostComponents...)
			} else {
				r.SubResources = append(r.SubResources, subResource)
			}
		}

		qty := int64(0)
		if a.InstanceCount != nil {
			qty = *a.InstanceCount
		}
		schema.MultiplyQuantities(r, decimal.NewFromInt(qty))
	}

	return r
}
