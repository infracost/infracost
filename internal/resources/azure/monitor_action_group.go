package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strconv"
)

// MonitorActionGroup struct represents an Azure Monitor Action Group.
//
// Resource information: https://learn.microsoft.com/en-us/azure/azure-monitor/alerts/action-groups
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/monitor/
type MonitorActionGroup struct {
	Address string
	Region  string

	MonthlySMSMessages map[string]float64 `infracost_usage:"monthly_sms_messages"`
	MonthlyVoiceCalls  map[string]float64 `infracost_usage:"monthly_voice_calls"`
}

func (r *MonitorActionGroup) CoreType() string {
	return "MonitorActionGroup"
}

func (r *MonitorActionGroup) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "monthly_sms_messages",
			ValueType:    schema.KeyValueMap,
			DefaultValue: map[string]float64{"country_code_1": 0},
		},
		{
			Key:          "monthly_voice_calls",
			ValueType:    schema.KeyValueMap,
			DefaultValue: map[string]float64{"country_code_1": 0},
		},
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

	// SMS messages
	smsCostComponents := []*schema.CostComponent{}
	smsCountryCodes, smsCallUsage := r.mapCountryCodesToQuantity(r.MonthlySMSMessages)
	for _, countryCode := range smsCountryCodes {
		smsCostComponents = append(smsCostComponents, r.smsMessageCostComponent(countryCode, smsCallUsage[countryCode]))
	}
	if len(smsCostComponents) > 0 {
		subResources = append(subResources, &schema.Resource{
			Name:           "SMS messages",
			CostComponents: smsCostComponents,
		})
	}

	// Voice calls
	voiceCallCostComponents := []*schema.CostComponent{}
	voiceCallCountryCodes, voiceCallUsage := r.mapCountryCodesToQuantity(r.MonthlyVoiceCalls)
	for _, countryCode := range voiceCallCountryCodes {
		voiceCallCostComponents = append(voiceCallCostComponents, r.voiceCallsCostComponent(countryCode, voiceCallUsage[countryCode]))
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

var countryCodeRegex = regexp.MustCompile(`^country_code_(\d+)$`)

func (r *MonitorActionGroup) mapCountryCodesToQuantity(usageMap map[string]float64) ([]int, map[int]float64) {
	countryCodes := make([]int, 0, len(usageMap))
	countryCodeToQuantity := make(map[int]float64, len(usageMap))

	for k, v := range usageMap {
		if match := countryCodeRegex.FindStringSubmatch(k); match != nil {
			code, _ := strconv.Atoi(match[1])
			countryCodes = append(countryCodes, code)
			countryCodeToQuantity[code] = v
		} else {
			log.Warnf("Unrecognized country code key %s, must match country_code_(\\d+)", k)
		}
	}

	sort.Ints(countryCodes)

	return countryCodes, countryCodeToQuantity
}

func (r *MonitorActionGroup) smsMessageCostComponent(countryCode int, quantity float64) *schema.CostComponent {
	var startUsageAmount string
	if countryCode == 1 {
		startUsageAmount = "100" // the first 10 US calls are free
	} else {
		startUsageAmount = "0"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Country code %d", countryCode),
		Unit:            "messages",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(quantity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(fmt.Sprintf("SMS Country Code %d", countryCode))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
		},
	}
}

func (r *MonitorActionGroup) voiceCallsCostComponent(countryCode int, quantity float64) *schema.CostComponent {
	var meterName string
	var startUsageAmount string
	if countryCode == 1 {
		meterName = "Voice Calls"
		startUsageAmount = "10" // the first 10 US calls are free
	} else {
		meterName = fmt.Sprintf("Voice Calls Voice Call Country Code %d", countryCode)
		startUsageAmount = "0"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Country code %d", countryCode),
		Unit:            "calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(quantity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Voice Calls")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsageAmount),
		},
	}
}
