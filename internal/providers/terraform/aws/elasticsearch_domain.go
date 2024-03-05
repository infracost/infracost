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
		ClusterInstanceType:           d.Get("cluster_config.0.instance_type").String(),
		EBSEnabled:                    d.Get("ebs_options.0.ebs_enabled").Bool(),
		EBSVolumeType:                 d.Get("ebs_options.0.volume_type").String(),
		ClusterDedicatedMasterEnabled: d.Get("cluster_config.0.dedicated_master_enabled").Bool(),
		ClusterDedicatedMasterType:    d.Get("cluster_config.0.dedicated_master_type").String(),
		ClusterWarmEnabled:            d.Get("cluster_config.0.warm_enabled").Bool(),
		ClusterWarmType:               d.Get("cluster_config.0.warm_type").String(),
	}

	if !d.IsEmpty("cluster_config.0.instance_count") {
		r.ClusterInstanceCount = intPtr(d.Get("cluster_config.0.instance_count").Int())
	}

	if !d.IsEmpty("ebs_options.0.volume_size") {
		r.EBSVolumeSize = floatPtr(d.Get("ebs_options.0.volume_size").Float())
	}

	if !d.IsEmpty("ebs_options.0.iops") {
		r.EBSIOPS = floatPtr(d.Get("ebs_options.0.iops").Float())
	}

	if !d.IsEmpty("ebs_options.0.throughput") {
		r.EBSThroughput = floatPtr(d.Get("ebs_options.0.throughput").Float())
	}

	if !d.IsEmpty("cluster_config.0.dedicated_master_count") {
		r.ClusterDedicatedMasterCount = intPtr(d.Get("cluster_config.0.dedicated_master_count").Int())
	}

	if !d.IsEmpty("cluster_config.0.warm_count") {
		r.ClusterWarmCount = intPtr(d.Get("cluster_config.0.warm_count").Int())
	}
	return r
}
