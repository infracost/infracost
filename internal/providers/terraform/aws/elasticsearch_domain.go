package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElasticsearchDomainRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticsearch_domain",
		RFunc: NewElasticsearchDomain,
	}
}
func NewElasticsearchDomain(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ElasticsearchDomain{Address: strPtr(d.Address), ClusterConfig0WarmType: strPtr(d.Get("cluster_config.0.warm_type").String()), Region: strPtr(d.Get("region").String()), ClusterConfig0WarmCount: intPtr(d.Get("cluster_config.0.warm_count").Int())}
	if !d.IsEmpty("cluster_config.0.instance_type") {
		r.ClusterConfig0InstanceType = strPtr(d.Get("cluster_config.0.instance_type").String())
	}
	if !d.IsEmpty("cluster_config.0.instance_count") {
		r.ClusterConfig0InstanceCount = intPtr(d.Get("cluster_config.0.instance_count").Int())
	}
	if !d.IsEmpty("cluster_config.0.dedicated_master_enabled") {
		r.ClusterConfig0DedicatedMasterEnabled = boolPtr(d.Get("cluster_config.0.dedicated_master_enabled").Bool())
	}
	if !d.IsEmpty("cluster_config.0.warm_enabled") {
		r.ClusterConfig0WarmEnabled = boolPtr(d.Get("cluster_config.0.warm_enabled").Bool())
	}
	if !d.IsEmpty("ebs_options.0.ebs_enabled") {
		r.EbsOptions0EbsEnabled = boolPtr(d.Get("ebs_options.0.ebs_enabled").Bool())
	}
	if !d.IsEmpty("ebs_options.0.volume_size") {
		r.EbsOptions0VolumeSize = floatPtr(d.Get("ebs_options.0.volume_size").Float())
	}
	if !d.IsEmpty("ebs_options.0.volume_type") {
		r.EbsOptions0VolumeType = strPtr(d.Get("ebs_options.0.volume_type").String())
	}
	if !d.IsEmpty("ebs_options.0.iops") {
		r.EbsOptions0Iops = floatPtr(d.Get("ebs_options.0.iops").Float())
	}
	if !d.IsEmpty("cluster_config.0.dedicated_master_type") {
		r.ClusterConfig0DedicatedMasterType = strPtr(d.Get("cluster_config.0.dedicated_master_type").String())
	}
	if !d.IsEmpty("cluster_config.0.dedicated_master_count") {
		r.ClusterConfig0DedicatedMasterCount = intPtr(d.Get("cluster_config.0.dedicated_master_count").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
