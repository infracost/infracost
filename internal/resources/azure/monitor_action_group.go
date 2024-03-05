package azure

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MonitorActionGroup struct represents an Azure Monitor Action Group.
//
// Resource information: https://learn.microsoft.com/en-us/azure/azure-monitor/alerts/action-groups
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/monitor/
type MonitorActionGroup struct {
	Address string
	Region  string

	EmailReceivers                  int
	ITSMEventReceivers              int
	PushNotificationReceivers       int
	SecureWebHookReceivers          int
	WebHookReceivers                int
	SMSReceiversByCountryCode       map[int]int
	VoiceCallReceiversByCountryCode map[int]int

	MonthlyNotifications *int64 `infracost_usage:"monthly_notifications"`
}

func (r *MonitorActionGroup) CoreType() string {
	return "MonitorActionGroup"
}

func (r *MonitorActionGroup) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_notifications", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData
// It uses the `infracost_usage` struct tags to populate data.
func (r *MonitorActionGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from the struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MonitorActionGroup) BuildResource() *schema.Resource {
	subResources := []*schema.Resource{}
	costComponents := []*schema.CostComponent{}

	if r.EmailReceivers > 0 {
		costComponents = append(costComponents, r.emailCostComponent(r.EmailReceivers, r.MonthlyNotifications))
	}
	if r.ITSMEventReceivers > 0 {
		costComponents = append(costComponents, r.ITSMEventCostComponent(r.ITSMEventReceivers, r.MonthlyNotifications))
	}
	if r.PushNotificationReceivers > 0 {
		costComponents = append(costComponents, r.pushNotificationCostComponent(r.PushNotificationReceivers, r.MonthlyNotifications))
	}
	if r.SecureWebHookReceivers > 0 {
		costComponents = append(costComponents, r.secureWebHookCostComponent(r.SecureWebHookReceivers, r.MonthlyNotifications))
	}
	if r.WebHookReceivers > 0 {
		costComponents = append(costComponents, r.webHookCostComponent(r.WebHookReceivers, r.MonthlyNotifications))
	}

	// SMS messages
	smsCostComponents := []*schema.CostComponent{}
	for _, countryCode := range r.getSortedKeys(r.SMSReceiversByCountryCode) {
		smsCostComponents = append(smsCostComponents, r.smsMessageCostComponent(countryCode, r.SMSReceiversByCountryCode[countryCode], r.MonthlyNotifications))
	}
	if len(smsCostComponents) > 0 {
		subResources = append(subResources, &schema.Resource{
			Name:           "SMS messages",
			CostComponents: smsCostComponents,
		})
	}

	// Voice calls
	voiceCallCostComponents := []*schema.CostComponent{}
	for _, countryCode := range r.getSortedKeys(r.VoiceCallReceiversByCountryCode) {
		voiceCallCostComponents = append(voiceCallCostComponents, r.voiceCallsCostComponent(countryCode, r.VoiceCallReceiversByCountryCode[countryCode], r.MonthlyNotifications))
	}
	if len(voiceCallCostComponents) > 0 {
		subResources = append(subResources, &schema.Resource{
			Name:           "Voice calls",
			CostComponents: voiceCallCostComponents,
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func (r *MonitorActionGroup) emailCostComponent(count int, quantity *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Email notifications (%d)", count),
		Unit:            "emails",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Emails")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("1000"),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) ITSMEventCostComponent(count int, quantity *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("ITSM connectors (%d)", count),
		Unit:            "events",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Notifications")},
				{Key: "meterName", Value: strPtr("Notifications ITSM Connector Create/Update Event")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("1000"),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) pushNotificationCostComponent(count int, quantity *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Push notifications (%d)", count),
		Unit:            "notifications",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Notifications")},
				{Key: "meterName", Value: strPtr("Notifications Push Notification")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("1000"),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) secureWebHookCostComponent(count int, quantity *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Secure web hooks (%d)", count),
		Unit:            "notifications",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Notifications")},
				{Key: "meterName", Value: strPtr("Notifications Secure web hook")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("100"),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) webHookCostComponent(count int, quantity *int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Web hooks (%d)", count),
		Unit:            "notifications",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Notifications")},
				{Key: "meterName", Value: strPtr("Notifications Web hook")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("10000"),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) smsMessageCostComponent(countryCode int, count int, quantity *int64) *schema.CostComponent {
	var startUsageAmount string
	if countryCode == 1 {
		startUsageAmount = "100" // the first 10 US calls are free
	} else {
		startUsageAmount = "0"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Country code %d (%d)", countryCode, count),
		Unit:            "messages",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(fmt.Sprintf("SMS Country Code %d", countryCode))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) voiceCallsCostComponent(countryCode int, count int, quantity *int64) *schema.CostComponent {
	var meterName string
	if countryCode == 1 {
		meterName = "Voice Calls"
	} else {
		meterName = fmt.Sprintf("Voice Calls Voice Call Country Code %d", countryCode)
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Country code %d (%d)", countryCode, count),
		Unit:            "calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: r.monthlyQuantity(count, quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Voice Calls")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("10"),
		},
		UsageBased: true,
	}
}

func (r *MonitorActionGroup) monthlyQuantity(count int, quantity *int64) *decimal.Decimal {
	if quantity == nil {
		return nil
	}

	return decimalPtr(decimal.NewFromInt(int64(count) * *quantity))
}

func (r *MonitorActionGroup) getSortedKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	return keys
}

func (r *MonitorActionGroup) normalizedRegion() *string {
	if r.Region == "global" {
		return strPtr("Global")
	}
	return strPtr(r.Region)
}
