package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMemoryDBSubnetGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_memorydb_subnet_group",
		CoreRFunc: NewMemoryDBSubnetGroup,
	}
}

func NewMemoryDBSubnetGroup(d *schema.ResourceData) schema.CoreResource {
	r := &aws.MemoryDBSubnetGroup{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
