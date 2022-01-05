package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getPubSubSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_pubsub_subscription",
		RFunc: NewPubsubSubscription,
	}
}
func NewPubsubSubscription(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.PubsubSubscription{Address: strPtr(d.Address)}
	r.PopulateUsage(u)
	return r.BuildResource()
}
