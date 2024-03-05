package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// DataFactoryIntegrationRuntimeSelfHosted struct represents Data Factory's
// Self-hosted runtime.
//
// Resource information: https://azure.microsoft.com/en-us/services/data-factory/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/data-factory/data-pipeline/
type DataFactoryIntegrationRuntimeSelfHosted struct {
	Address string
	Region  string

	// "usage" args
	MonthlyOrchestrationRuns *int64 `infracost_usage:"monthly_orchestration_runs"`
}

func (r *DataFactoryIntegrationRuntimeSelfHosted) CoreType() string {
	return "DataFactoryIntegrationRuntimeSelfHosted"
}

func (r *DataFactoryIntegrationRuntimeSelfHosted) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_orchestration_runs", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the DataFactoryIntegrationRuntimeSelfHosted.
// It uses the `infracost_usage` struct tags to populate data into the DataFactoryIntegrationRuntimeSelfHosted.
func (r *DataFactoryIntegrationRuntimeSelfHosted) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid DataFactoryIntegrationRuntimeSelfHosted struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *DataFactoryIntegrationRuntimeSelfHosted) BuildResource() *schema.Resource {
	runtimeFilter := "Self Hosted"

	costComponents := []*schema.CostComponent{
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
