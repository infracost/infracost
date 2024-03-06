package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudwatchLogGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudwatch_log_group",
		CoreRFunc: NewCloudwatchLogGroup,
	}
}
func NewCloudwatchLogGroup(d *schema.ResourceData) schema.CoreResource {
	r := &aws.CloudwatchLogGroup{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
