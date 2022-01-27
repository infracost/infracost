package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSNSTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sns_topic",
		RFunc: NewSNSTopic,
	}
}

func NewSNSTopic(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SNSTopic{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
