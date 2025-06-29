package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElasticsearchDomainRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_elasticsearch_domain",
		CoreRFunc: newSearchDomain,
	}
}
func newSearchDomain(d *schema.ResourceData) schema.CoreResource {
	r := &aws.SearchDomain{
		Address:                       d.Address,
		Region:                        d.Get("region").String(),
		ClusterInstanceType:           d.GetChild("cluster_config").Get("instance_type").String(),
		EBSEnabled:                    d.GetChild("ebs_options").Get("ebs_enabled").Bool(),
		EBSVolumeType:                 d.GetChild("ebs_options").Get("volume_type").String(),
		ClusterDedicatedMasterEnabled: d.GetChild("cluster_config").Get("dedicated_master_enabled").Bool(),
		ClusterDedicatedMasterType:    d.GetChild("cluster_config").Get("dedicated_master_type").String(),
		ClusterWarmEnabled:            d.GetChild("cluster_config").Get("warm_enabled").Bool(),
		ClusterWarmType:               d.GetChild("cluster_config").Get("warm_type").String(),
	}

	if !d.IsEmpty("cluster_config.instance_count") {
		r.ClusterInstanceCount = intPtr(d.GetChild("cluster_config").Get("instance_count").Int())
	}

	if !d.IsEmpty("ebs_options.volume_size") {
		r.EBSVolumeSize = floatPtr(d.GetChild("ebs_options").Get("volume_size").Float())
	}

	if !d.IsEmpty("ebs_options.iops") {
		r.EBSIOPS = floatPtr(d.GetChild("ebs_options").Get("iops").Float())
	}

	if !d.IsEmpty("ebs_options.throughput") {
		r.EBSThroughput = floatPtr(d.GetChild("ebs_options").Get("throughput").Float())
	}

	if !d.IsEmpty("cluster_config.dedicated_master_count") {
		r.ClusterDedicatedMasterCount = intPtr(d.GetChild("cluster_config").Get("dedicated_master_count").Int())
	}

	if !d.IsEmpty("cluster_config.warm_count") {
		r.ClusterWarmCount = intPtr(d.GetChild("cluster_config").Get("warm_count").Int())
	}

	return r
}
