package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSNSTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_sns_topic",
		RFunc:           NewSNSTopic,
		ReferenceAttributes: []string{"awsSnsTopicSubscription.topicArn"},
	}
}

func NewSNSTopic(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if d.GetBoolOrDefault("fifoTopic", false) {
		r := &aws.SNSFIFOTopic{
			Address:       d.Address,
			Region:        d.Get("region").String(),
			Subscriptions: int64(len(d.References("awsSnsTopicSubscription.topicArn"))),
		}

	r.PopulateUsage(u)
	return r.BuildResource()
	}

	r := &aws.SNSTopic{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
