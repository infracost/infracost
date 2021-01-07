package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudwatchMetricAlarmRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_metric_alarm",
		RFunc: NewCloudwatchMetricAlarm,
	}
}

func NewCloudwatchMetricAlarm(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)

	if len(d.Get("metric_query").Array()) > 0 {
		costComponents = append(costComponents, cloudWatchMetricQuery(d, region, "alarm metrics"))
	} else {
		costComponents = append(costComponents, cloudWatchMetricName(d, region, "alarm metrics"))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func cloudwatchMetricAlarmCostComponent(name string, unit string, region string, alarmType string, quantity decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonCloudWatch"),
			ProductFamily: strPtr("Alarm"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "alarmType", Value: strPtr(alarmType)},
			},
		},
	}
}

func cloudWatchMetricQuery(d *schema.ResourceData, region string, costComponentUnit string) *schema.CostComponent {
	var quantity decimal.Decimal
	var name string
	var alarmType string
	var anomalyDetection string

	metricCount := 0

	for _, metric := range d.Get("metric_query.#.metric").Array() {
		if len(metric.Array()) > 0 {
			metricCount++
		}
	}

	quantity = decimal.NewFromInt(int64(metricCount))

	if checkAnomlyDetection(d) {
		quantity = quantity.Mul(decimal.NewFromInt(3))
		anomalyDetection = " anomaly detection"
		costComponentUnit = "alarms"
	}

	for _, query := range d.Get("metric_query").Array() {
		if len(query.Get("metric").Array()) == 0 {
			continue
		}

		for _, m := range query.Get("metric").Array() {
			if m.Get("period").Exists() {
				if calcMetricResolution(decimal.NewFromInt(m.Get("period").Int())) {
					name = fmt.Sprintf("%s%s", "Standard resolution", anomalyDetection)
					alarmType = "Standard"
				} else {
					name = fmt.Sprintf("%s%s", "High resolution", anomalyDetection)
					alarmType = "High Resolution"
				}

				return cloudwatchMetricAlarmCostComponent(name, costComponentUnit, region, alarmType, quantity)
			}
		}
	}

	return nil
}

func cloudWatchMetricName(d *schema.ResourceData, region string, costComponentUnit string) *schema.CostComponent {
	var name string
	var alarmType string

	if calcMetricResolution(decimal.NewFromInt(d.Get("period").Int())) {
		name = "Standard resolution"
		alarmType = "Standard"
	} else {
		name = "High resolution"
		alarmType = "High Resolution"
	}

	return cloudwatchMetricAlarmCostComponent(name, costComponentUnit, region, alarmType, decimal.NewFromInt(1))
}

func calcMetricResolution(metricPeriod decimal.Decimal) bool {
	return metricPeriod.Div(decimal.NewFromInt(60)).GreaterThanOrEqual(decimal.NewFromInt(1))
}

func checkAnomlyDetection(d *schema.ResourceData) bool {
	switch d.Get("comparison_operator").String() {
	case "LessThanLowerOrGreaterThanUpperThreshold", "LessThanLowerThreshold", "GreaterThanUpperThreshold":
		return true
	}
	return false
}
