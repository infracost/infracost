package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getPubSubSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_pubsub_subscription",
		RFunc: NewPubSubSubscription,
	}
}

func NewPubSubSubscription(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.PubSubSubscription{
		Address: d.Address,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
