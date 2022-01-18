package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getPubSubTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_pubsub_topic",
		RFunc: NewPubsubTopic,
	}
}
func NewPubsubTopic(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.PubsubTopic{Address: strPtr(d.Address)}
	r.PopulateUsage(u)
	return r.BuildResource()
}
