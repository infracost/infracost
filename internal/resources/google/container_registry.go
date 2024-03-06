package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

type ContainerRegistry struct {
	Address                     string
	Region                      string
	Location                    string
	StorageClass                string
	StorageGB                   *float64                             `infracost_usage:"storage_gb"`
	MonthlyClassAOperations     *int64                               `infracost_usage:"monthly_class_a_operations"`
	MonthlyClassBOperations     *int64                               `infracost_usage:"monthly_class_b_operations"`
	MonthlyEgressDataTransferGB *ContainerRegistryNetworkEgressUsage `infracost_usage:"monthly_egress_data_transfer_gb"`
}

func (r *ContainerRegistry) CoreType() string {
	return "ContainerRegistry"
}

func (r *ContainerRegistry) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_class_a_operations", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_class_b_operations", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_data_retrieval_gb", ValueType: schema.Float64, DefaultValue: 0},
		{
			Key:          "monthly_egress_data_transfer_gb",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "monthly_egress_data_transfer_gb", Items: ContainerRegistryNetworkEgressUsageSchema},
		},
	}
}

func (r *ContainerRegistry) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ContainerRegistry) BuildResource() *schema.Resource {
	if r.MonthlyEgressDataTransferGB == nil {
		r.MonthlyEgressDataTransferGB = &ContainerRegistryNetworkEgressUsage{}
	}
	region := r.Region
	components := []*schema.CostComponent{
		dataStorageCostComponent(r.Location, r.StorageClass, r.StorageGB),
	}

	components = append(components, operationsCostComponents(r.StorageClass, r.MonthlyClassAOperations, r.MonthlyClassBOperations)...)

	r.MonthlyEgressDataTransferGB.Region = region
	r.MonthlyEgressDataTransferGB.Address = "Network egress"
	r.MonthlyEgressDataTransferGB.PrefixName = "Data transfer"
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		SubResources: []*schema.Resource{
			r.MonthlyEgressDataTransferGB.BuildResource(),
		}, UsageSchema: r.UsageSchema(),
	}
}
