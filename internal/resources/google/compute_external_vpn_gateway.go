package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

type ComputeExternalVPNGateway struct {
	Address string
	Region  string

	MonthlyEgressDataTransferGB *ComputeExternalVPNGatewayNetworkEgressUsage `infracost_usage:"monthly_egress_data_transfer_gb"`
}

func (r *ComputeExternalVPNGateway) CoreType() string {
	return "ComputeExternalVPNGateway"
}

func (r *ComputeExternalVPNGateway) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:       "monthly_egress_data_transfer_gb",
			ValueType: schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "monthly_egress_data_transfer_gb",
				Items: ComputeExternalVPNGatewayNetworkEgressUsageSchema},
		},
	}
}

func (r *ComputeExternalVPNGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
	if r.MonthlyEgressDataTransferGB == nil {
		r.MonthlyEgressDataTransferGB = &ComputeExternalVPNGatewayNetworkEgressUsage{}
	}
}

func (r *ComputeExternalVPNGateway) BuildResource() *schema.Resource {
	region := r.Region
	r.MonthlyEgressDataTransferGB.Region = region
	r.MonthlyEgressDataTransferGB.Address = "Network egress"
	r.MonthlyEgressDataTransferGB.PrefixName = "IPSec traffic"
	return &schema.Resource{
		Name: r.Address,
		SubResources: []*schema.Resource{
			r.MonthlyEgressDataTransferGB.BuildResource(),
		}, UsageSchema: r.UsageSchema(),
	}
}
