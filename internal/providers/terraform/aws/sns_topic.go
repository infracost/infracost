package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSNSTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_sns_topic",
		RFunc:               NewSNSTopic,
		ReferenceAttributes: []string{"aws_sns_topic_subscription.topic_arn"},
	}
}

func NewSNSTopic(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if d.GetBoolOrDefault("fifo_topic", false) {
		r := &aws.SNSFIFOTopic{
			Address:       d.Address,
			Region:        d.Get("region").String(),
			Subscriptions: int64(len(d.References("aws_sns_topic_subscription.topic_arn"))),
		}

		r.PopulateUsage(u)
		return r.BuildResource()
	}

	r := &aws.SNSTopic{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
