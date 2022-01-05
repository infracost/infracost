package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetLoggingBillingAccountSinkRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_logging_billing_account_sink",
		RFunc: NewLoggingBillingAccountSink,
	}
}

func NewLoggingBillingAccountSink(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var loggingData *decimal.Decimal
	if u != nil && u.Get("monthly_logging_data_gb").Exists() {
		loggingData = decimalPtr(decimal.NewFromInt(u.Get("monthly_logging_data_gb").Int()))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: loggingCostComponent(loggingData),
	}
}
