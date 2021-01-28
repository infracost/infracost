package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestRoute53Record(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_zone" "zone1" {
			name = "example.com"
		}

		resource "aws_route53_record" "standard" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "standard.example.com"
			type    = "A"
			ttl     = "300"
			records = ["10.0.0.1"]
		}

		resource "aws_route53_record" "geo" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "geo.example.com"
			type    = "A"
			ttl     = "300"
			records = ["10.0.0.2"]
			geolocation_routing_policy {
				continent = "NA"
			}
		}

		resource "aws_route53_record" "latency" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "latency.example.com"
			type    = "A"
			ttl     = "300"
			records = ["10.0.0.3"]
			latency_routing_policy {
				region = "us-west-1"
			}
		}

		resource "aws_elb" "elb1" {
			availability_zones = ["us-east-1c"]
			listener {
				instance_port     = 80
				instance_protocol = "http"
				lb_port           = 80
				lb_protocol       = "http"
			}
		}

		resource "aws_route53_record" "alias1" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "alias1.example.com"
			type    = "A"

			alias {
				name                   = aws_elb.elb1.dns_name
				zone_id                = aws_elb.elb1.zone_id
				evaluate_target_health = true
			}
		}

		resource "aws_route53_record" "alias2" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "alias2.example.com"
			type    = "A"

			alias {
				name                   = aws_route53_record.standard.name
				zone_id                = aws_route53_record.standard.zone_id
				evaluate_target_health = true
			}
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_route53_zone.zone1",
			SkipCheck: true,
		},
		{
			Name:      "aws_elb.elb1",
			SkipCheck: true,
		},
		{
			Name: "aws_route53_record.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard queries",
					PriceHash:        "c07c948553cc6492cc58c7b53b8dfdf2-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_route53_record.geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Geo DNS queries",
					PriceHash:        "1565af203c9c0e9a59815a64b9c484d0-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_route53_record.latency",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Latency based routing queries",
					PriceHash:        "82e2ac0a19cdd4c54fea556c3f8c3892-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_route53_record.alias2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard queries",
					PriceHash:        "c07c948553cc6492cc58c7b53b8dfdf2-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
