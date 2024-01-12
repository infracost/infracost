package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetFlowLogRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "states.aws.ec2.flow_log.present",
		CoreRFunc: func(d *schema.ResourceData) schema.CoreResource {
			return schema.BlankCoreResource{
				Name: d.Address,
				Type: d.Type,
			}
		},
	}
}
