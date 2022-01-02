package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDocDBClusterSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster_snapshot",
		RFunc: NewDocdbClusterSnapshot,
	}

}
func NewDocdbClusterSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DocdbClusterSnapshot{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
