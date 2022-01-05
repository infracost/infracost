package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSNSTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sns_topic",
		RFunc: NewSnsTopic,
	}
}
func NewSnsTopic(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SnsTopic{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
