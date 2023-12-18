package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudHSMv2HSMRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudhsm_v2_hsm",
		CoreRFunc: newCloudHSMv2HSM,
	}
}

func newCloudHSMv2HSM(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	return &aws.CloudHSMv2HSM{
		Address: d.Address,
		Region:  region,
	}
}
