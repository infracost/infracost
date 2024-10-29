package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type CloudwatchMetricAlarm struct {
	Address            string
	Region             string
	ComparisonOperator string
	Metrics            int64
	Period             int64
}

func (r *CloudwatchMetricAlarm) CoreType() string {
	return "CloudwatchMetricAlarm"
}

func (r *CloudwatchMetricAlarm) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *CloudwatchMetricAlarm) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudwatchMetricAlarm) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, r.cloudwatchMetricAlarmCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *CloudwatchMetricAlarm) cloudwatchMetricAlarmCostComponent() *schema.CostComponent {
	var name string
	var alarmType string
	var anomalyDetection string

	unit := "alarm metrics"
	quantity := decimal.NewFromInt(int64(r.Metrics))

	if r.checkAnomalyDetection() {
		quantity = quantity.Mul(decimal.NewFromInt(3))
		anomalyDetection = " anomaly detection"
		unit = "alarms"
	}

	if r.CalcMetricResolution(decimal.NewFromInt(r.Period)) {
		name = fmt.Sprintf("%s%s", "Standard resolution", anomalyDetection)
		alarmType = "Standard"
	} else {
		name = fmt.Sprintf("%s%s", "High resolution", anomalyDetection)
		alarmType = "High Resolution"
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Alarm"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "alarmType", ValueRegex: regexPtr(alarmType)},
				{Key: "usagetype", ValueRegex: regexPtr("AlarmMonitorUsage$")},
			},
		},
	}
}

func (r *CloudwatchMetricAlarm) CalcMetricResolution(metricPeriod decimal.Decimal) bool {
	return metricPeriod.Div(decimal.NewFromInt(60)).GreaterThanOrEqual(decimal.NewFromInt(1))
}

func (r *CloudwatchMetricAlarm) checkAnomalyDetection() bool {
	switch r.ComparisonOperator {
	case "LessThanLowerOrGreaterThanUpperThreshold", "LessThanLowerThreshold", "GreaterThanUpperThreshold":
		return true
	}
	return false
}
