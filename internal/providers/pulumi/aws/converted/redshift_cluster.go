package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRedshiftClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_redshift_cluster",
		RFunc: NewRedshiftCluster,
	}
}

func NewRedshiftCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RedshiftCluster{
		Address:  d.Address,
		Region:   d.Get("region").String(),
		NodeType: d.Get("nodeType").String(),
	}

	if !d.IsEmpty("number_of_nodes") {
		r.Nodes = intPtr(d.Get("numberOfNodes").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
