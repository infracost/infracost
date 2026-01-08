package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMSKClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_msk_cluster",
		CoreRFunc:           NewMSKCluster,
		ReferenceAttributes: []string{"aws_appautoscaling_target.resource_id"},
	}
}
func NewMSKCluster(d *schema.ResourceData) schema.CoreResource {
	refs := d.References("aws_appautoscaling_target.resource_id")
	targets := make([]*aws.AppAutoscalingTarget, 0, len(refs))
	for _, ref := range refs {
		targets = append(targets, newAppAutoscalingTarget(ref, ref.UsageData))
	}

	var brokerEBSVolumeSize int64
	if d.Get("broker_node_group_info.0.ebs_volume_size").Exists() {
		// terraform-provider-aws v4
		brokerEBSVolumeSize = d.Get("broker_node_group_info.0.ebs_volume_size").Int()
	} else {
		// terraform-provider-aws v5
		brokerEBSVolumeSize = d.Get("broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size").Int()
	}

	r := &aws.MSKCluster{
		Address:                 d.Address,
		Region:                  d.Get("region").String(),
		BrokerNodes:             d.Get("number_of_broker_nodes").Int(),
		BrokerNodeInstanceType:  d.Get("broker_node_group_info.0.instance_type").String(),
		BrokerNodeEBSVolumeSize: brokerEBSVolumeSize,
		AppAutoscalingTarget:    targets,
	}
	return r
}
