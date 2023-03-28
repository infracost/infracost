package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// MonitorScheduledQueryRulesAlert struct represents Azure Monitor Log Alert Rules,
// aka Scheduled Query Rules.
//
// Resource information:
//
//	 https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/monitor_scheduled_query_rules_alert_v2
//		https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/monitor_scheduled_query_rules_alert
//
// Pricing information: https://azure.microsoft.com/en-in/pricing/details/monitor/
type MonitorScheduledQueryRulesAlert struct {
	Address string
	Region  string

	Enabled          bool
	TimeSeriesCount  int64
	FrequencyMinutes int64
}

// CoreType returns the name of this resource type
func (r *MonitorScheduledQueryRulesAlert) CoreType() string {
	return "MonitorScheduledQueryRulesAlert"
}

// UsageSchema defines a list which represents the usage schema of MonitorScheduledQueryRulesAlert.
func (r *MonitorScheduledQueryRulesAlert) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the MonitorScheduledQueryRulesAlert.
// It uses the `infracost_usage` struct tags to populate data into the MonitorScheduledQueryRulesAlert.
func (r *MonitorScheduledQueryRulesAlert) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid MonitorScheduledQueryRulesAlert struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MonitorScheduledQueryRulesAlert) BuildResource() *schema.Resource {
	if !r.Enabled {
		return &schema.Resource{
			Name: r.Address,
		}
	}

	var normalizedFrequency int
	switch {
	case r.FrequencyMinutes >= 15:
		normalizedFrequency = 15
	case r.FrequencyMinutes >= 10:
		normalizedFrequency = 10
	case r.FrequencyMinutes >= 5:
		normalizedFrequency = 5
	default:
		normalizedFrequency = 1
	}

	costComponents := []*schema.CostComponent{
		r.logAlertMonitoringCostComponent(normalizedFrequency),
	}

	if r.TimeSeriesCount > 1 {
		costComponents = append(costComponents, r.logAlertAdditionalTimeSeriesCostComponent(normalizedFrequency, r.TimeSeriesCount-1))
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *MonitorScheduledQueryRulesAlert) logAlertMonitoringCostComponent(normalizedFrequency int) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Log alerts monitoring (%d minute frequency)", normalizedFrequency),
		Unit:            "rule",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(fmt.Sprintf("Alerts System Log Monitored at %d Minute Frequency", normalizedFrequency))},
			},
		},
	}
}

func (r *MonitorScheduledQueryRulesAlert) logAlertAdditionalTimeSeriesCostComponent(normalizedFrequency int, additionalCount int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Additional time-series monitoring (%d minute frequency)", normalizedFrequency),
		Unit:            "time-series",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(additionalCount)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(fmt.Sprintf("Alerts Resource Monitored at %d Minute Frequency", normalizedFrequency))},
			},
		},
	}
}

func (r *MonitorScheduledQueryRulesAlert) normalizedRegion() *string {
	if r.Region == "global" {
		return strPtr("Global")
	}
	return strPtr(r.Region)
}
