package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetRedshiftClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_redshift_cluster",
		RFunc: NewRedshiftCluster,
	}
}
func NewRedshiftCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RedshiftCluster{Address: strPtr(d.Address), NodeType: strPtr(d.Get("node_type").String()), Region: strPtr(d.Get("region").String())}
	if d.Get("number_of_nodes").Exists() && d.Get("number_of_nodes").Type != gjson.Null {
		r.NumberOfNodes = intPtr(d.Get("number_of_nodes").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
