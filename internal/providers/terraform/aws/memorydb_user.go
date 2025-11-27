package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMemoryDBUserRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_memorydb_user",
		CoreRFunc: NewMemoryDBUser,
	}
}

func NewMemoryDBUser(d *schema.ResourceData) schema.CoreResource {
	r := &aws.MemoryDBUser{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
