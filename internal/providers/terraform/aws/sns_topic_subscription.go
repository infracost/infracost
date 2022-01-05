package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSNSTopicSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sns_topic_subscription",
		RFunc: NewSnsTopicSubscription,
		Notes: []string{
			"SMS and mobile push not yet supported.",
		},
	}
}
func NewSnsTopicSubscription(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SnsTopicSubscription{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), Protocol: strPtr(d.Get("protocol").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
