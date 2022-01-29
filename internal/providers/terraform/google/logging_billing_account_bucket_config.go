package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getLoggingBillingAccountBucketConfigRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_logging_billing_account_bucket_config",
		RFunc: NewLoggingBillingAccountBucketConfig,
	}
}

func NewLoggingBillingAccountBucketConfig(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.Logging{
		Address: d.Address,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
