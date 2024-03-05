package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudwatchDashboardRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudwatch_dashboard",
		CoreRFunc: NewCloudwatchDashboard,
	}
}
func NewCloudwatchDashboard(d *schema.ResourceData) schema.CoreResource {
	r := &aws.CloudwatchDashboard{
		Address: d.Address,
	}
	return r
}
