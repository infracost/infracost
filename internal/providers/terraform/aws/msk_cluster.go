package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMSKClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_msk_cluster",
		RFunc: NewMskCluster,
	}
}
func NewMskCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.MskCluster{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), NumberOfBrokerNodes: intPtr(d.Get("number_of_broker_nodes").Int()), BrokerNodeGroupInfo0InstanceType: strPtr(d.Get("broker_node_group_info.0.instance_type").String()), BrokerNodeGroupInfo0EbsVolumeSize: intPtr(d.Get("broker_node_group_info.0.ebs_volume_size").Int())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
