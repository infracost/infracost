package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

type ComputeVPNGateway struct {
	Address string
	Region  string

	MonthlyEgressDataTransferGB *ComputeVPNGatewayNetworkEgressUsage `infracost_usage:"monthly_egress_data_transfer_gb"`
}

func (r *ComputeVPNGateway) CoreType() string {
	return "ComputeVPNGateway"
}

func (r *ComputeVPNGateway) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "monthly_egress_data_transfer_gb",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "monthly_egress_data_transfer_gb", Items: ComputeVPNGatewayNetworkEgressUsageSchema},
		},
	}
}

func (r *ComputeVPNGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeVPNGateway) BuildResource() *schema.Resource {
	if r.MonthlyEgressDataTransferGB == nil {
		r.MonthlyEgressDataTransferGB = &ComputeVPNGatewayNetworkEgressUsage{}
	}
	region := r.Region
	r.MonthlyEgressDataTransferGB.Region = region
	r.MonthlyEgressDataTransferGB.Address = "Network egress"
	r.MonthlyEgressDataTransferGB.PrefixName = "IPSec traffic"
	return &schema.Resource{
		Name: r.Address,
		SubResources: []*schema.Resource{
			r.MonthlyEgressDataTransferGB.BuildResource(),
		},
		UsageSchema: r.UsageSchema(),
	}
}
