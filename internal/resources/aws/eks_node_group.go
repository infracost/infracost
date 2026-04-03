package aws

import (
	"context"
	"math"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
)

type EKSNodeGroup struct {
	// "required" args that can't really be missing.
	Address     string
	Region      string
	Name        string
	ClusterName string

	InstanceType   string
	PurchaseOption string
	DiskSize       float64

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

func (a *EKSNodeGroup) CoreType() string {
	return "EKSNodeGroup"
}

func (a *EKSNodeGroup) UsageSchema() []*schema.UsageItem {
	return EKSNodeGroupUsageSchema
}

var EKSNodeGroupUsageSchema = append([]*schema.UsageItem{
	{Key: "instances", DefaultValue: 0, ValueType: schema.Int64},
}, InstanceUsageSchema...)

func (a *EKSNodeGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)

	// The usage keys for Launch Template are specified on the EKS Node Groupresource
	if a.LaunchTemplate != nil {
		resources.PopulateArgsWithUsage(a.LaunchTemplate, u)
	}
}

// getUsageSchemaWithDefaultInstanceCount is a temporary hack to make --sync-usage-file use the node group's "desired_size"
// as the default value for the "instances" usage param.  Without this, --sync-usage-file sets instances=0 causing the
// costs for the node group to be $0.  This can be removed when --sync-usage-file creates the usage file with usgage keys
// commented out by default.
func (a *EKSNodeGroup) getUsageSchemaWithDefaultInstanceCount() []*schema.UsageItem {
	if a.InstanceCount == nil || *a.InstanceCount == 0 {
		return EKSNodeGroupUsageSchema
	}

	usageSchema := make([]*schema.UsageItem, 0, len(EKSNodeGroupUsageSchema))
	for _, u := range EKSNodeGroupUsageSchema {
		if u.Key == "instances" {
			usageSchema = append(usageSchema, &schema.UsageItem{Key: "instances", DefaultValue: intVal(a.InstanceCount), ValueType: schema.Int64})
		} else {
			usageSchema = append(usageSchema, u)
		}
	}
	return usageSchema
}

func (a *EKSNodeGroup) BuildResource() *schema.Resource {
	r := &schema.Resource{
		Name:        a.Address,
		UsageSchema: a.getUsageSchemaWithDefaultInstanceCount(),
	}

	var estimateInstanceQualities schema.EstimateFunc

	// The EKS Node Group resource either has the instance attributes inline or a reference to a Launch Template.
	// If it has a reference to a Launch Template we create generic resources for that and add add it as a subresource
	// of the EKS Node Group resource.
	if a.LaunchTemplate != nil {
		lt := a.LaunchTemplate.BuildResource()
		// If the Launch Template returns nil it is not supported so the Autoscaling Group should also return nil
		if lt == nil {
			return nil
		}
		r.SubResources = append(r.SubResources, lt)
		estimateInstanceQualities = lt.EstimateUsage
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
			Size:    intPtr(int64(a.DiskSize)),
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
		estimateInstanceQualities = instanceResource.EstimateUsage

		qty := int64(0)
		if a.InstanceCount != nil {
			qty = *a.InstanceCount
		}
		schema.MultiplyQuantities(r, decimal.NewFromInt(qty))
	}

	estimate := func(ctx context.Context, u map[string]any) error {
		err := estimateInstanceQualities(ctx, u)
		if err != nil {
			return err
		}

		// By default (no LaunchTemplate), instances use Amazon Linux 2 AMI."
		// c.f. https://docs.aws.amazon.com/eks/latest/userguide/managed-node-groups.html
		if _, ok := u["operating_system"]; !ok {
			u["operating_system"] = "linux"
		}

		if a.ClusterName != "" && a.Name != "" {
			// Sum the counts from the EKS Node Group's Autoscaling Groups
			asgNames, err := aws.EKSGetNodeGroupAutoscalingGroups(ctx, a.Region, a.ClusterName, a.Name)
			if err != nil {
				return err
			}
			count := float64(0.0)

			if len(asgNames) > 0 {
				for _, asgName := range asgNames {
					asgCount, err := aws.AutoscalingGetInstanceCount(ctx, a.Region, asgName)
					if err != nil {
						return err
					}
					count += asgCount
				}
			}

			if count > 0 {
				u["instances"] = int64(math.Round(count))
			}
		}
		return nil
	}
	r.EstimateUsage = estimate

	return r
}
