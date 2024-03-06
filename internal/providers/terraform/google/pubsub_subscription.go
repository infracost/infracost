package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getPubSubSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_pubsub_subscription",
		CoreRFunc: NewPubSubSubscription,
	}
}

func NewPubSubSubscription(d *schema.ResourceData) schema.CoreResource {
	r := &google.PubSubSubscription{
		Address: d.Address,
	}

	return r
}
