package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSNSTopicSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_sns_topic_subscription",
		CoreRFunc: NewSNSTopicSubscription,
		Notes: []string{
			"DEPRECATED.  Set subscription usage on aws_sns_topic instead.",
		},
		ReferenceAttributes: []string{"topic_arn"},
	}
}

func NewSNSTopicSubscription(d *schema.ResourceData) schema.CoreResource {
	r := &aws.SNSTopicSubscription{
		Address:  d.Address,
		Region:   d.Get("region").String(),
		Protocol: d.Get("protocol").String(),
	}
	return r
}
