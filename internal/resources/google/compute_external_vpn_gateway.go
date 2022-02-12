package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

type ComputeExternalVpnGateway struct {
	Address string
	Region  string

	MonthlyEgressDataTransferGB *NetworkEgressUsage `infracost_usage:"monthly_egress_data_transfer_gb"`
}

var ComputeExternalVpnGatewayUsageSchema = []*schema.UsageItem{
	{
		Key:          "monthly_egress_data_transfer_gb",
		ValueType:    schema.SubResourceUsage,
		DefaultValue: &usage.ResourceUsage{Name: "monthly_egress_data_transfer_gb", Items: NetworkEgressUsageSchema},
	},
}

func (r *ComputeExternalVpnGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeExternalVpnGateway) BuildResource() *schema.Resource {
	region := r.Region
	return &schema.Resource{
		Name: r.Address,
		SubResources: []*schema.Resource{
			r.MonthlyEgressDataTransferGB.networkEgress(region, "Network egress", "IPSec traffic", ComputeExternalVPNGateway),
		}, UsageSchema: ComputeExternalVpnGatewayUsageSchema,
	}
}
