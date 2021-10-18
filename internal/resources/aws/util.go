package aws

import (
	"math"
	"reflect"
	"regexp"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

var (
	underscore = regexp.MustCompile(`_`)
)

func strPtr(s string) *string {
	return &s
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intPtr(i int64) *int64 {
	return &i
}

func intVal(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func intPtrToDecimalPtr(i *int64) *decimal.Decimal {
	if i == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromInt(*i))
}

func floatPtrToDecimalPtr(f *float64) *decimal.Decimal {
	if f == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromFloat(*f))
}

func asGiB(i int64) int64 {
	if i == 0 {
		return 0
	}
	i /= (1024 * 1024 * 1024)
	if i == 0 {
		return 1
	}
	return i
}

func ceil64(f float64) int64 {
	return int64(math.Ceil(f))
}

func stringInSlice(slice []string, s string) bool {
	for _, b := range slice {
		if b == s {
			return true
		}
	}
	return false
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

// RegionsUsage is a reusable type that represents a usage cost map.
// This can be used in resources that define a usage parameter that's
// changed on a per-region basis. e.g.
//
// monthly_data_processed_gb:
//   us_gov_west_1: 188
//   us_east_1: 78
//
// can be handled by adding a usage cost property to your resource like so:
//
// type MyResource struct {
//    ...
//    MonthlyDataProcessedGB *RegionsUsage `infracost_usage:"monthly_processed_gb"`
// }
type RegionsUsage struct {
	UsGovWest1   *int64 `infracost_usage:"us_gov_west_1"`
	UsGovEast1   *int64 `infracost_usage:"us_gov_east_1"`
	UsEast1      *int64 `infracost_usage:"us_east_1"`
	UsEast2      *int64 `infracost_usage:"us_east_2"`
	UsWest1      *int64 `infracost_usage:"us_west_1"`
	UsWest2      *int64 `infracost_usage:"us_west_2"`
	UsWest2Lax1  *int64 `infracost_usage:"us_west_2_lax_1"`
	CaCentral1   *int64 `infracost_usage:"ca_central_1"`
	CnNorth1     *int64 `infracost_usage:"cn_north_1"`
	CnNorthwest1 *int64 `infracost_usage:"cn_northwest_1"`
	EuCentral1   *int64 `infracost_usage:"eu_central_1"`
	EuWest1      *int64 `infracost_usage:"eu_west_1"`
	EuWest2      *int64 `infracost_usage:"eu_west_2"`
	EuSouth1     *int64 `infracost_usage:"eu_south_1"`
	EuWest3      *int64 `infracost_usage:"eu_west_3"`
	EuNorth1     *int64 `infracost_usage:"eu_north_1"`
	ApEast1      *int64 `infracost_usage:"ap_east_1"`
	ApNortheast1 *int64 `infracost_usage:"ap_northeast_1"`
	ApNortheast2 *int64 `infracost_usage:"ap_northeast_2"`
	ApNortheast3 *int64 `infracost_usage:"ap_northeast_3"`
	ApSoutheast1 *int64 `infracost_usage:"ap_southeast_1"`
	ApSoutheast2 *int64 `infracost_usage:"ap_southeast_2"`
	ApSouth1     *int64 `infracost_usage:"ap_south_1"`
	MeSouth1     *int64 `infracost_usage:"me_south_1"`
	SaEast1      *int64 `infracost_usage:"sa_east_1"`
	AfSouth1     *int64 `infracost_usage:"af_south_1"`
}

// RegionUsage defines a hard definition in the regions map.
type RegionUsage struct {
	Key   string
	Value int64
}

// Values returns RegionUsage as a slice which can be iterated over
// to create cost components. The keys of the regions returned have
// their underscores replaced with hypens so they can be used in
// product filters and cost lookups.
func (r RegionsUsage) Values() []RegionUsage {
	s := reflect.ValueOf(r)
	t := reflect.TypeOf(r)

	var regions []RegionUsage
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if f.IsNil() {
			continue
		}

		regions = append(regions, RegionUsage{
			Key:   underscore.ReplaceAllString(t.Field(i).Tag.Get("infracost_usage"), "-"),
			Value: *f.Interface().(*int64),
		})
	}

	return regions
}

// RegionUsageSchema is the schema representation of the RegionsUsage type.
// This can be used as a schema.SubResourceUsage to define a structure that's
// commonly used with data transfer usage. e.g:
//
// 		monthly_data_transfer_out_gb:
//			us_gov_west_1: 122
//			ca_central_1: 99
//
// See DirectoryServiceDirectory for an example usage.
var RegionUsageSchema = []*schema.UsageItem{
	{Key: "us_gov_west_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "us_gov_east_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "us_east_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "us_east_2", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "us_west_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "us_west_2", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "us_west_2_lax_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ca_central_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "cn_north_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "cn_northwest_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "eu_central_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "eu_west_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "eu_west_2", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "eu_south_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "eu_west_3", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "eu_north_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_east_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_northeast_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_northeast_2", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_northeast_3", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_southeast_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_southeast_2", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "ap_south_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "me_south_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "sa_east_1", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "af_south_1", DefaultValue: 0, ValueType: schema.Int64},
}
