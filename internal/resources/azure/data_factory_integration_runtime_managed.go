package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// DataFactoryIntegrationRuntimeManaged struct represents Data Factory's Managed
// VNET integration runtime.
//
// Resource information: https://azure.microsoft.com/en-us/services/data-factory/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/data-factory/data-pipeline/
type DataFactoryIntegrationRuntimeManaged struct {
	Address string
	Region  string

	Instances       int64
	InstanceType    string
	Enterprise      bool
	LicenseIncluded bool

	// "usage" args
	MonthlyOrchestrationRuns *int64 `infracost_usage:"monthly_orchestration_runs"`
}

func (r *DataFactoryIntegrationRuntimeManaged) CoreType() string {
	return "DataFactoryIntegrationRuntimeManaged"
}

func (r *DataFactoryIntegrationRuntimeManaged) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_orchestration_runs", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the DataFactoryIntegrationRuntimeManaged.
// It uses the `infracost_usage` struct tags to populate data into the DataFactoryIntegrationRuntimeManaged.
func (r *DataFactoryIntegrationRuntimeManaged) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid DataFactoryIntegrationRuntimeManaged struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *DataFactoryIntegrationRuntimeManaged) BuildResource() *schema.Resource {
	runtimeFilter := "Azure Managed VNET"

	// SSIS and Managed runtime resources share the same compute configuration.
	// Terraform provider has deprecated Managed VNET runtime resource in favor of
	// SSIS one.
	ssis := DataFactoryIntegrationRuntimeAzureSSIS{
		Address:         r.Address,
		Region:          r.Region,
		Enterprise:      r.Enterprise,
		LicenseIncluded: r.LicenseIncluded,
		Instances:       r.Instances,
		InstanceType:    r.InstanceType,
	}

	costComponents := []*schema.CostComponent{
		ssis.computeCostComponent(),
		dataFactoryOrchestrationCostComponent(r.Region, runtimeFilter, r.MonthlyOrchestrationRuns),
		dataFactoryDataMovementCostComponent(r.Region, runtimeFilter),
		dataFactoryPipelineCostComponent(r.Region, runtimeFilter),
		dataFactoryExternalPipelineCostComponent(r.Region, runtimeFilter),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
