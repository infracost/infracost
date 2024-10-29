package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlobalAcceleratorRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_globalaccelerator_accelerator",
		CoreRFunc: newGlobalAccelerator,
	}
}

func newGlobalAccelerator(d *schema.ResourceData) schema.CoreResource {

	r := &aws.GlobalAccelerator{
		Address: d.Address,
	}

	return r
}
