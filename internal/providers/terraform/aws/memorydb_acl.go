package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMemoryDBACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_memorydb_acl",
		CoreRFunc: NewMemoryDBACL,
	}
}

func NewMemoryDBACL(d *schema.ResourceData) schema.CoreResource {
	r := &aws.MemoryDBACL{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
