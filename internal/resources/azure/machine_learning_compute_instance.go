package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MachineLearningComputeInstance struct represents a Azure Machine Learning Compute Instance.
//
// These use the same pricing as Azure Linux Virtual Machines.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/machine-learning/#overview
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/machine-learning/
type MachineLearningComputeInstance struct {
	Address      string
	Region       string
	InstanceType string
	MonthlyHours *float64 `infracost_usage:"monthly_hrs"`
}

// CoreType returns the name of this resource type
func (r *MachineLearningComputeInstance) CoreType() string {
	return "MachineLearningComputeInstance"
}

// UsageSchema defines a list which represents the usage schema of MachineLearningComputeInstance.
func (r *MachineLearningComputeInstance) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the MachineLearningComputeInstance.
// It uses the `infracost_usage` struct tags to populate data into the MachineLearningComputeInstance.
func (r *MachineLearningComputeInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid MachineLearningComputeInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MachineLearningComputeInstance) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		linuxVirtualMachineCostComponent(r.Region, r.InstanceType, r.MonthlyHours),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
