package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetCloudwatchEventBusItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_event_bus",
		RFunc: NewCloudwatchEventBusItem,
	}
}
func NewCloudwatchEventBusItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.CloudwatchEventBusItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
