package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetCloudwatchMetricAlarmRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.cloudwatch.metric_alarm.present",
		RFunc: newCloudwatchMetricAlarm,
	}
}
func newCloudwatchMetricAlarm(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	comparisonOperator := d.Get("comparison_operator").String()

	var metricCount int64
	var period int64

	if len(d.Get("metrics").Array()) > 0 {
		metricCount = 0
		for _, metric := range d.Get("metrics").Array() {
			metrics := metric.Array()

			if len(metrics) == 0 {
				continue
			}

			metricCount++

			for _, m := range metrics {
				if period == 0 && m.Get("Period").Exists() {
					period = m.Get("Period").Int()
				}
			}
		}
	} else {
		metricCount = 1
		period = d.Get("period").Int()
	}

	r := &aws.CloudwatchMetricAlarm{
		Address:            d.Address,
		Region:             region,
		ComparisonOperator: comparisonOperator,
		Metrics:            metricCount,
		Period:             period,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
