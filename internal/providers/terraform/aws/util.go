package aws

import "github.com/shopspring/decimal"

var defaultVolumeSize = 8

func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

var regionMapping = map[string]string{
	"us-gov-west-1":   "AWS GovCloud (US-West)",
	"us-gov-east-1":   "AWS GovCloud (US-East)",
	"us-east-1":       "US East (N. Virginia)",
	"us-east-2":       "US East (Ohio)",
	"us-west-1":       "US West (N. California)",
	"us-west-2":       "US West (Oregon)",
	"us-west-2-lax-1": "US West (Los Angeles)",
	"ca-central-1":    "Canada (Central)",
	"cn-north-1":      "China (Beijing)",
	"cn-northwest-1":  "China (Ningxia)",
	"eu-central-1":    "EU (Frankfurt)",
	"eu-west-1":       "EU (Ireland)",
	"eu-west-2":       "EU (London)",
	"eu-south-1":      "EU (Milan)",
	"eu-west-3":       "EU (Paris)",
	"eu-north-1":      "EU (Stockholm)",
	"ap-east-1":       "Asia Pacific (Hong Kong)",
	"ap-northeast-1":  "Asia Pacific (Tokyo)",
	"ap-northeast-2":  "Asia Pacific (Seoul)",
	"ap-northeast-3":  "Asia Pacific (Osaka)",
	"ap-southeast-1":  "Asia Pacific (Singapore)",
	"ap-southeast-2":  "Asia Pacific (Sydney)",
	"ap-south-1":      "Asia Pacific (Mumbai)",
	"me-south-1":      "Middle East (Bahrain)",
	"sa-east-1":       "South America (Sao Paulo)",
	"af-south-1":      "Africa (Cape Town)",
}
