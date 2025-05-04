package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMemoryDBSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_memorydb_snapshot",
		CoreRFunc: NewMemoryDBSnapshot,
	}
}

func NewMemoryDBSnapshot(d *schema.ResourceData) schema.CoreResource {
	r := &aws.MemoryDBSnapshot{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
