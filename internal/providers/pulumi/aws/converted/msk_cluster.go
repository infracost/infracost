package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMSKClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_msk_cluster",
		RFunc:           NewMSKCluster,
		ReferenceAttributes: []string{"awsAppautoscalingTarget.resourceId"},
	}
}
func NewMSKCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	targets := []*aws.AppAutoscalingTarget{}
	for _, ref := range d.References("awsAppautoscalingTarget.resourceId") {
		targets = append(targets, newAppAutoscalingTarget(ref, ref.UsageData))
	}

	var brokerEBSVolumeSize int64
	if d.Get("brokerNodeGroupInfo.0.ebsVolumeSize").Exists() {
		// terraform-provider-aws v4
		brokerEBSVolumeSize = d.Get("brokerNodeGroupInfo.0.ebsVolumeSize").Int()
	} else {
		// terraform-provider-aws v5
		brokerEBSVolumeSize = d.Get("brokerNodeGroupInfo.0.storageInfo.0.ebsStorageInfo.0.volumeSize").Int()
	}

	r := &aws.MSKCluster{
		Address:                 d.Address,
		Region:                  d.Get("region").String(),
		BrokerNodes:             d.Get("numberOfBrokerNodes").Int(),
		BrokerNodeInstanceType:  d.Get("brokerNodeGroupInfo.0.instanceType").String(),
		BrokerNodeEBSVolumeSize: brokerEBSVolumeSize,
		AppAutoscalingTarget:    targets,
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
