package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestCloudWatchMetricAlarm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_cloudwatch_metric_alarm" "standard" {
          alarm_name                = "terraform-test-alarm"
          comparison_operator       = "GreaterThanOrEqualToThreshold"
          evaluation_periods        = "2"
          metric_name               = "CPUUtilization"
          namespace                 = "AWS/EC2"
          period                    = "120"
          statistic                 = "Average"
          threshold                 = "80"
          alarm_description         = "This metric monitors ec2 cpu utilization"
          insufficient_data_actions = []
        }

        resource "aws_cloudwatch_metric_alarm" "high" {
          alarm_name                = "terraform-test-alarm"
          comparison_operator       = "GreaterThanOrEqualToThreshold"
          evaluation_periods        = "1"
          metric_name               = "CPUUtilization"
          namespace                 = "AWS/EC2"
          period                    = "10"
          statistic                 = "Average"
          threshold                 = "80"
          alarm_description         = "This metric monitors ec2 cpu utilization"
          insufficient_data_actions = []
        }

         resource "aws_cloudwatch_metric_alarm" "expression" {
           alarm_name                = "terraform-test-expression"
           comparison_operator       = "GreaterThanOrEqualToThreshold"
           evaluation_periods        = "2"
           threshold                 = "10"
           alarm_description         = "Request error rate has exceeded 10%"
           insufficient_data_actions = []

           metric_query {
             id          = "e1"
             expression  = "m2/m1*100"
             label       = "Error Rate"
             return_data = "true"
           }

           metric_query {
             id = "m1"

             metric {
               metric_name = "RequestCount"
               namespace   = "AWS/ApplicationELB"
               period      = "120"
               stat        = "Sum"
               unit        = "Count"

               dimensions = {
                 LoadBalancer = "app/web"
               }
             }
           }

           metric_query {
             id = "m2"

             metric {
               metric_name = "HTTPCode_ELB_5XX_Count"
               namespace   = "AWS/ApplicationELB"
               period      = "120"
               stat        = "Sum"
               unit        = "Count"

               dimensions = {
                 LoadBalancer = "app/web"
               }
             }
           }
         }

         resource "aws_cloudwatch_metric_alarm" "anomaly_detection" {
           alarm_name                = "terraform-test-foobar"
           comparison_operator       = "GreaterThanUpperThreshold"
           evaluation_periods        = "2"
           threshold_metric_id       = "e1"
           alarm_description         = "This metric monitors ec2 cpu utilization"
           insufficient_data_actions = []

           metric_query {
             id          = "e1"
             expression  = "ANOMALY_DETECTION_BAND(m1)"
             label       = "CPUUtilization (Expected)"
             return_data = "true"
           }

           metric_query {
             id          = "m1"
             return_data = "true"
             metric {
               metric_name = "CPUUtilization"
               namespace   = "AWS/EC2"
               period      = "30"
               stat        = "Average"
               unit        = "Count"

               dimensions = {
                 InstanceId = "i-abc123"
               }
             }
           }
         }

`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_metric_alarm.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard resolution",
					PriceHash:        "a84a546af7ac086f8ba1e9e80712d14e-062aaa56dee17578c258c2b52f349952",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_cloudwatch_metric_alarm.high",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "High resolution",
					PriceHash:        "2909875a1b2145747d5ee321e71aeaf1-062aaa56dee17578c258c2b52f349952",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_cloudwatch_metric_alarm.expression",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard resolution",
					PriceHash:        "a84a546af7ac086f8ba1e9e80712d14e-062aaa56dee17578c258c2b52f349952",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
		{
			Name: "aws_cloudwatch_metric_alarm.anomaly_detection",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "High resolution anomaly detection",
					PriceHash:        "2909875a1b2145747d5ee321e71aeaf1-062aaa56dee17578c258c2b52f349952",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
