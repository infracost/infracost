package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRedshiftClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_redshift_cluster",
		CoreRFunc: NewRedshiftCluster,
	}
}

func NewRedshiftCluster(d *schema.ResourceData) schema.CoreResource {
	r := &aws.RedshiftCluster{
		Address:  d.Address,
		Region:   d.Get("region").String(),
		NodeType: d.Get("node_type").String(),
	}

	if !d.IsEmpty("number_of_nodes") {
		r.Nodes = intPtr(d.Get("number_of_nodes").Int())
	}
	return r
}
