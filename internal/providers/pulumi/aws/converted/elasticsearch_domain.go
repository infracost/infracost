package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElasticsearchDomainRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_elasticsearch_domain",
		RFunc: newSearchDomain,
	}
}
func newSearchDomain(d *schema.ResourceData) schema.CoreResource {
	r := &aws.SearchDomain{
		Address:                       d.Address,
		Region:                        d.Get("region").String(),
		ClusterInstanceType:           d.Get("clusterConfig.0.instanceType").String(),
		EBSEnabled:                    d.Get("ebsOptions.0.ebsEnabled").Bool(),
		EBSVolumeType:                 d.Get("ebsOptions.0.volumeType").String(),
		ClusterDedicatedMasterEnabled: d.Get("clusterConfig.0.dedicatedMasterEnabled").Bool(),
		ClusterDedicatedMasterType:    d.Get("clusterConfig.0.dedicatedMasterType").String(),
		ClusterWarmEnabled:            d.Get("clusterConfig.0.warmEnabled").Bool(),
		ClusterWarmType:               d.Get("clusterConfig.0.warmType").String(),
	}

	if !d.IsEmpty("cluster_config.0.instance_count") {
		r.ClusterInstanceCount = intPtr(d.Get("clusterConfig.0.instanceCount").Int())
	}

	if !d.IsEmpty("ebs_options.0.volume_size") {
		r.EBSVolumeSize = floatPtr(d.Get("ebsOptions.0.volumeSize").Float())
	}

	if !d.IsEmpty("ebs_options.0.iops") {
		r.EBSIOPS = floatPtr(d.Get("ebsOptions.0.iops").Float())
	}

	if !d.IsEmpty("ebs_options.0.throughput") {
		r.EBSThroughput = floatPtr(d.Get("ebsOptions.0.throughput").Float())
	}

	if !d.IsEmpty("cluster_config.0.dedicated_master_count") {
		r.ClusterDedicatedMasterCount = intPtr(d.Get("clusterConfig.0.dedicatedMasterCount").Int())
	}

	if !d.IsEmpty("cluster_config.0.warm_count") {
		r.ClusterWarmCount = intPtr(d.Get("clusterConfig.0.warmCount").Int())
	}
	return r
}
