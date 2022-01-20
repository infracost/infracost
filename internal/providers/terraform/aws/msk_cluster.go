package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMSKClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_msk_cluster",
		RFunc: NewMSKCluster,
	}
}
func NewMSKCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.MSKCluster{
		Address:                 d.Address,
		Region:                  d.Get("region").String(),
		BrokerNodes:             d.Get("number_of_broker_nodes").Int(),
		BrokerNodeInstanceType:  d.Get("broker_node_group_info.0.instance_type").String(),
		BrokerNodeEBSVolumeSize: d.Get("broker_node_group_info.0.ebs_volume_size").Int(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
