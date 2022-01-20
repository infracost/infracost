package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRedshiftClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_redshift_cluster",
		RFunc: NewRedshiftCluster,
	}
}

func NewRedshiftCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RedshiftCluster{
		Address:  d.Address,
		Region:   d.Get("region").String(),
		NodeType: d.Get("node_type").String(),
	}

	if !d.IsEmpty("number_of_nodes") {
		r.Nodes = intPtr(d.Get("number_of_nodes").Int())
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
