package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudwatchMetricAlarmRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudwatch_metric_alarm",
		CoreRFunc: newCloudwatchMetricAlarm,
	}
}
func newCloudwatchMetricAlarm(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	comparisonOperator := d.Get("comparison_operator").String()

	var metricCount int64
	var period int64

	if len(d.Get("metric_query").Array()) > 0 {
		metricCount = 0
		for _, metric := range d.Get("metric_query.#.metric").Array() {
			metrics := metric.Array()

			if len(metrics) == 0 {
				continue
			}

			metricCount++

			for _, m := range metrics {
				if period == 0 && m.Get("period").Exists() {
					period = m.Get("period").Int()
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
	return r
}
