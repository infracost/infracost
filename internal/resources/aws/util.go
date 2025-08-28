package aws

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

var (
	underscore = regexp.MustCompile(`_`)
	vendorName = strPtr("aws")
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

// nolint:deadcode,unused
func regexPtr(regex string) *string {
	return strPtr(fmt.Sprintf("/%s/i", regex))
}

func intPtr(i int64) *int64 {
	return &i
}

func floatPtr(i float64) *float64 {
	return &i
}

func intVal(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func floatVal(i *float64) float64 {
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

func sliceOfKeysFromMap[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// RegionMapping is a helpful conversion map that changes
// aws region name to the name commonly used in pricing filters.
var RegionMapping = map[string]string{
	"us-gov-west-1":   "AWS GovCloud (US-West)",
	"us-gov-east-1":   "AWS GovCloud (US-East)",
	"us-east-1":       "US East (N. Virginia)",
	"us-east-2":       "US East (Ohio)",
	"us-west-1":       "US West (N. California)",
	"us-west-2":       "US West (Oregon)",
	"us-west-2-lax-1": "US West (Los Angeles)",
	"ca-central-1":    "Canada (Central)",
	"ca-west-1":       "Canada West (Calgary)",
	"mx-central-1":    "Mexico (Central)",
	"cn-north-1":      "China (Beijing)",
	"cn-northwest-1":  "China (Ningxia)",
	"eu-central-1":    "EU (Frankfurt)",
	"eu-central-2":    "EU (Zurich)",
	"eu-west-1":       "EU (Ireland)",
	"eu-west-2":       "EU (London)",
	"eu-south-1":      "EU (Milan)",
	"eu-south-2":      "EU (Spain)",
	"eu-west-3":       "EU (Paris)",
	"eu-north-1":      "EU (Stockholm)",
	"il-central-1":    "Israel (Tel Aviv)",
	"ap-east-1":       "Asia Pacific (Hong Kong)",
	"ap-east-2":       "Asia Pacific (Taipei)",
	"ap-northeast-1":  "Asia Pacific (Tokyo)",
	"ap-northeast-2":  "Asia Pacific (Seoul)",
	"ap-northeast-3":  "Asia Pacific (Osaka)",
	"ap-southeast-1":  "Asia Pacific (Singapore)",
	"ap-southeast-2":  "Asia Pacific (Sydney)",
	"ap-southeast-3":  "Asia Pacific (Jakarta)",
	"ap-southeast-4":  "Asia Pacific (Melbourne)",
	"ap-southeast-5":  "Asia Pacific (Malaysia)",
	"ap-southeast-7":  "Asia Pacific (Thailand)",
	"ap-south-1":      "Asia Pacific (Mumbai)",
	"ap-south-2":      "Asia Pacific (Hyderabad)",
	"me-central-1":    "Middle East (UAE)",
	"me-south-1":      "Middle East (Bahrain)",
	"sa-east-1":       "South America (Sao Paulo)",
	"af-south-1":      "Africa (Cape Town)",
}

// RegionCodeMapping helps to find region's abbreviated code for a more granular
// filtering when resources may have multiple products for the same region.
var RegionCodeMapping = map[string]string{
	"ap-southeast-1": "APS1",
}

// RegionsUsage is a reusable type that represents a usage cost map.
// This can be used in resources that define a usage parameter that's
// changed on a per-region basis. e.g.
//
// monthly_data_processed_gb:
//
//	us_gov_west_1: 188
//	us_east_1: 78
//
// can be handled by adding a usage cost property to your resource like so:
//
//	type MyResource struct {
//	   ...
//	   MonthlyDataProcessedGB *RegionsUsage `infracost_usage:"monthly_processed_gb"`
//	}
type RegionsUsage struct {
	USGovWest1   *float64 `infracost_usage:"us_gov_west_1"`
	USGovEast1   *float64 `infracost_usage:"us_gov_east_1"`
	USEast1      *float64 `infracost_usage:"us_east_1"`
	USEast2      *float64 `infracost_usage:"us_east_2"`
	USWest1      *float64 `infracost_usage:"us_west_1"`
	USWest2      *float64 `infracost_usage:"us_west_2"`
	USWest2Lax1  *float64 `infracost_usage:"us_west_2_lax_1"`
	CACentral1   *float64 `infracost_usage:"ca_central_1"`
	CAWest1      *float64 `infracost_usage:"ca_west_1"`
	CNNorth1     *float64 `infracost_usage:"cn_north_1"`
	CNNorthwest1 *float64 `infracost_usage:"cn_northwest_1"`
	EUCentral1   *float64 `infracost_usage:"eu_central_1"`
	EUCentral2   *float64 `infracost_usage:"eu_central_2"`
	EUWest1      *float64 `infracost_usage:"eu_west_1"`
	EUWest2      *float64 `infracost_usage:"eu_west_2"`
	EUSouth1     *float64 `infracost_usage:"eu_south_1"`
	EUSouth2     *float64 `infracost_usage:"eu_south_2"`
	EUWest3      *float64 `infracost_usage:"eu_west_3"`
	EUNorth1     *float64 `infracost_usage:"eu_north_1"`
	ILCentral1   *float64 `infracost_usage:"il_central_1"`
	APEast1      *float64 `infracost_usage:"ap_east_1"`
	APEast2      *float64 `infracost_usage:"ap_east_2"`
	APNortheast1 *float64 `infracost_usage:"ap_northeast_1"`
	APNortheast2 *float64 `infracost_usage:"ap_northeast_2"`
	APNortheast3 *float64 `infracost_usage:"ap_northeast_3"`
	APSoutheast1 *float64 `infracost_usage:"ap_southeast_1"`
	APSoutheast2 *float64 `infracost_usage:"ap_southeast_2"`
	APSoutheast3 *float64 `infracost_usage:"ap_southeast_3"`
	APSoutheast4 *float64 `infracost_usage:"ap_southeast_4"`
	APSoutheast5 *float64 `infracost_usage:"ap_southeast_5"`
	APSoutheast7 *float64 `infracost_usage:"ap_southeast_7"`
	APSouth1     *float64 `infracost_usage:"ap_south_1"`
	APSouth2     *float64 `infracost_usage:"ap_south_2"`
	MECentral1   *float64 `infracost_usage:"me_central_1"`
	MESouth1     *float64 `infracost_usage:"me_south_1"`
	SAEast1      *float64 `infracost_usage:"sa_east_1"`
	AFSouth1     *float64 `infracost_usage:"af_south_1"`
}

// RegionUsage defines a hard definition in the regions map.
type RegionUsage struct {
	Key   string
	Value float64
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
			Value: *f.Interface().(*float64),
		})
	}

	return regions
}

// RegionUsageSchema is the schema representation of the RegionsUsage type.
// This can be used as a schema.SubResourceUsage to define a structure that's
// commonly used with data transfer usage. e.g:
//
//	monthly_data_transfer_out_gb:
//		us_gov_west_1: 122
//		ca_central_1: 99
//
// See DirectoryServiceDirectory for an example usage.
var RegionUsageSchema = []*schema.UsageItem{
	{Key: "us_gov_west_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "us_gov_east_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "us_east_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "us_east_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "us_west_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "us_west_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "us_west_2_lax_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ca_central_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ca_west_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "mx_central_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "cn_north_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "cn_northwest_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_central_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_central_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_west_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_west_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_south_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_south_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_west_3", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "eu_north_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "il_central_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_east_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_east_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_northeast_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_northeast_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_northeast_3", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_southeast_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_southeast_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_southeast_3", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_southeast_4", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_southeast_5", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_southeast_7", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_south_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "ap_south_2", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "me_central_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "me_south_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "sa_east_1", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "af_south_1", DefaultValue: 0, ValueType: schema.Float64},
}

func getBurstableInstanceFamily(matchPrefixes []string, instanceType string) string {
	for _, prefix := range matchPrefixes {
		if strings.HasPrefix(instanceType, prefix) {
			return prefix
		}
	}

	return ""
}

// this map was generated with:
// aws ec2 describe-instance-types | jq -r '[.InstanceTypes[] | "\"" + .InstanceType + "\": " + (.VCpuInfo.DefaultVCpus | tostring) + ","] | sort | .[]'
var InstanceTypeToVCPU = map[string]int64{
	"a1.2xlarge":       8,
	"a1.4xlarge":       16,
	"a1.large":         2,
	"a1.medium":        1,
	"a1.metal":         16,
	"a1.xlarge":        4,
	"c4.2xlarge":       8,
	"c4.4xlarge":       16,
	"c4.8xlarge":       36,
	"c4.large":         2,
	"c4.xlarge":        4,
	"c5.12xlarge":      48,
	"c5.18xlarge":      72,
	"c5.24xlarge":      96,
	"c5.2xlarge":       8,
	"c5.4xlarge":       16,
	"c5.9xlarge":       36,
	"c5.large":         2,
	"c5.metal":         96,
	"c5.xlarge":        4,
	"c5a.12xlarge":     48,
	"c5a.16xlarge":     64,
	"c5a.24xlarge":     96,
	"c5a.2xlarge":      8,
	"c5a.4xlarge":      16,
	"c5a.8xlarge":      32,
	"c5a.large":        2,
	"c5a.xlarge":       4,
	"c5ad.12xlarge":    48,
	"c5ad.16xlarge":    64,
	"c5ad.24xlarge":    96,
	"c5ad.2xlarge":     8,
	"c5ad.4xlarge":     16,
	"c5ad.8xlarge":     32,
	"c5ad.large":       2,
	"c5ad.xlarge":      4,
	"c5d.12xlarge":     48,
	"c5d.18xlarge":     72,
	"c5d.24xlarge":     96,
	"c5d.2xlarge":      8,
	"c5d.4xlarge":      16,
	"c5d.9xlarge":      36,
	"c5d.large":        2,
	"c5d.metal":        96,
	"c5d.xlarge":       4,
	"c5n.18xlarge":     72,
	"c5n.2xlarge":      8,
	"c5n.4xlarge":      16,
	"c5n.9xlarge":      36,
	"c5n.large":        2,
	"c5n.metal":        72,
	"c5n.xlarge":       4,
	"c6g.12xlarge":     48,
	"c6g.16xlarge":     64,
	"c6g.2xlarge":      8,
	"c6g.4xlarge":      16,
	"c6g.8xlarge":      32,
	"c6g.large":        2,
	"c6g.medium":       1,
	"c6g.metal":        64,
	"c6g.xlarge":       4,
	"c6gd.12xlarge":    48,
	"c6gd.16xlarge":    64,
	"c6gd.2xlarge":     8,
	"c6gd.4xlarge":     16,
	"c6gd.8xlarge":     32,
	"c6gd.large":       2,
	"c6gd.medium":      1,
	"c6gd.metal":       64,
	"c6gd.xlarge":      4,
	"c6gn.12xlarge":    48,
	"c6gn.16xlarge":    64,
	"c6gn.2xlarge":     8,
	"c6gn.4xlarge":     16,
	"c6gn.8xlarge":     32,
	"c6gn.large":       2,
	"c6gn.medium":      1,
	"c6gn.xlarge":      4,
	"c6i.12xlarge":     48,
	"c6i.16xlarge":     64,
	"c6i.24xlarge":     96,
	"c6i.2xlarge":      8,
	"c6i.32xlarge":     128,
	"c6i.4xlarge":      16,
	"c6i.8xlarge":      32,
	"c6i.large":        2,
	"c6i.metal":        128,
	"c6i.xlarge":       4,
	"d2.2xlarge":       8,
	"d2.4xlarge":       16,
	"d2.8xlarge":       36,
	"d2.xlarge":        4,
	"d3.2xlarge":       8,
	"d3.4xlarge":       16,
	"d3.8xlarge":       32,
	"d3.xlarge":        4,
	"g3.16xlarge":      64,
	"g3.4xlarge":       16,
	"g3.8xlarge":       32,
	"g3s.xlarge":       4,
	"g4ad.16xlarge":    64,
	"g4ad.2xlarge":     8,
	"g4ad.4xlarge":     16,
	"g4ad.8xlarge":     32,
	"g4ad.xlarge":      4,
	"g4dn.12xlarge":    48,
	"g4dn.16xlarge":    64,
	"g4dn.2xlarge":     8,
	"g4dn.4xlarge":     16,
	"g4dn.8xlarge":     32,
	"g4dn.metal":       96,
	"g4dn.xlarge":      4,
	"h1.16xlarge":      64,
	"h1.2xlarge":       8,
	"h1.4xlarge":       16,
	"h1.8xlarge":       32,
	"hpc6a.48xlarge":   96,
	"i2.2xlarge":       8,
	"i2.4xlarge":       16,
	"i2.8xlarge":       32,
	"i2.xlarge":        4,
	"i3.16xlarge":      64,
	"i3.2xlarge":       8,
	"i3.4xlarge":       16,
	"i3.8xlarge":       32,
	"i3.large":         2,
	"i3.metal":         72,
	"i3.xlarge":        4,
	"i3en.12xlarge":    48,
	"i3en.24xlarge":    96,
	"i3en.2xlarge":     8,
	"i3en.3xlarge":     12,
	"i3en.6xlarge":     24,
	"i3en.large":       2,
	"i3en.metal":       96,
	"i3en.xlarge":      4,
	"im4gn.16xlarge":   64,
	"im4gn.2xlarge":    8,
	"im4gn.4xlarge":    16,
	"im4gn.8xlarge":    32,
	"im4gn.large":      2,
	"im4gn.xlarge":     4,
	"inf1.24xlarge":    96,
	"inf1.2xlarge":     8,
	"inf1.6xlarge":     24,
	"inf1.xlarge":      4,
	"is4gen.2xlarge":   8,
	"is4gen.4xlarge":   16,
	"is4gen.8xlarge":   32,
	"is4gen.large":     2,
	"is4gen.medium":    1,
	"is4gen.xlarge":    4,
	"m4.10xlarge":      40,
	"m4.16xlarge":      64,
	"m4.2xlarge":       8,
	"m4.4xlarge":       16,
	"m4.large":         2,
	"m4.xlarge":        4,
	"m5.12xlarge":      48,
	"m5.16xlarge":      64,
	"m5.24xlarge":      96,
	"m5.2xlarge":       8,
	"m5.4xlarge":       16,
	"m5.8xlarge":       32,
	"m5.large":         2,
	"m5.metal":         96,
	"m5.xlarge":        4,
	"m5a.12xlarge":     48,
	"m5a.16xlarge":     64,
	"m5a.24xlarge":     96,
	"m5a.2xlarge":      8,
	"m5a.4xlarge":      16,
	"m5a.8xlarge":      32,
	"m5a.large":        2,
	"m5a.xlarge":       4,
	"m5ad.12xlarge":    48,
	"m5ad.16xlarge":    64,
	"m5ad.24xlarge":    96,
	"m5ad.2xlarge":     8,
	"m5ad.4xlarge":     16,
	"m5ad.8xlarge":     32,
	"m5ad.large":       2,
	"m5ad.xlarge":      4,
	"m5d.12xlarge":     48,
	"m5d.16xlarge":     64,
	"m5d.24xlarge":     96,
	"m5d.2xlarge":      8,
	"m5d.4xlarge":      16,
	"m5d.8xlarge":      32,
	"m5d.large":        2,
	"m5d.metal":        96,
	"m5d.xlarge":       4,
	"m5dn.12xlarge":    48,
	"m5dn.16xlarge":    64,
	"m5dn.24xlarge":    96,
	"m5dn.2xlarge":     8,
	"m5dn.4xlarge":     16,
	"m5dn.8xlarge":     32,
	"m5dn.large":       2,
	"m5dn.metal":       96,
	"m5dn.xlarge":      4,
	"m5n.12xlarge":     48,
	"m5n.16xlarge":     64,
	"m5n.24xlarge":     96,
	"m5n.2xlarge":      8,
	"m5n.4xlarge":      16,
	"m5n.8xlarge":      32,
	"m5n.large":        2,
	"m5n.metal":        96,
	"m5n.xlarge":       4,
	"m5zn.12xlarge":    48,
	"m5zn.2xlarge":     8,
	"m5zn.3xlarge":     12,
	"m5zn.6xlarge":     24,
	"m5zn.large":       2,
	"m5zn.metal":       48,
	"m5zn.xlarge":      4,
	"m6g.12xlarge":     48,
	"m6g.16xlarge":     64,
	"m6g.2xlarge":      8,
	"m6g.4xlarge":      16,
	"m6g.8xlarge":      32,
	"m6g.large":        2,
	"m6g.medium":       1,
	"m6g.metal":        64,
	"m6g.xlarge":       4,
	"m6gd.12xlarge":    48,
	"m6gd.16xlarge":    64,
	"m6gd.2xlarge":     8,
	"m6gd.4xlarge":     16,
	"m6gd.8xlarge":     32,
	"m6gd.large":       2,
	"m6gd.medium":      1,
	"m6gd.metal":       64,
	"m6gd.xlarge":      4,
	"m6i.12xlarge":     48,
	"m6i.16xlarge":     64,
	"m6i.24xlarge":     96,
	"m6i.2xlarge":      8,
	"m6i.32xlarge":     128,
	"m6i.4xlarge":      16,
	"m6i.8xlarge":      32,
	"m6i.large":        2,
	"m6i.metal":        128,
	"m6i.xlarge":       4,
	"mac1.metal":       12,
	"p2.16xlarge":      64,
	"p2.8xlarge":       32,
	"p2.xlarge":        4,
	"p3.16xlarge":      64,
	"p3.2xlarge":       8,
	"p3.8xlarge":       32,
	"p4d.24xlarge":     96,
	"r3.2xlarge":       8,
	"r3.4xlarge":       16,
	"r3.8xlarge":       32,
	"r3.large":         2,
	"r3.xlarge":        4,
	"r4.16xlarge":      64,
	"r4.2xlarge":       8,
	"r4.4xlarge":       16,
	"r4.8xlarge":       32,
	"r4.large":         2,
	"r4.xlarge":        4,
	"r5.12xlarge":      48,
	"r5.16xlarge":      64,
	"r5.24xlarge":      96,
	"r5.2xlarge":       8,
	"r5.4xlarge":       16,
	"r5.8xlarge":       32,
	"r5.large":         2,
	"r5.metal":         96,
	"r5.xlarge":        4,
	"r5a.12xlarge":     48,
	"r5a.16xlarge":     64,
	"r5a.24xlarge":     96,
	"r5a.2xlarge":      8,
	"r5a.4xlarge":      16,
	"r5a.8xlarge":      32,
	"r5a.large":        2,
	"r5a.xlarge":       4,
	"r5ad.12xlarge":    48,
	"r5ad.16xlarge":    64,
	"r5ad.24xlarge":    96,
	"r5ad.2xlarge":     8,
	"r5ad.4xlarge":     16,
	"r5ad.8xlarge":     32,
	"r5ad.large":       2,
	"r5ad.xlarge":      4,
	"r5b.12xlarge":     48,
	"r5b.16xlarge":     64,
	"r5b.24xlarge":     96,
	"r5b.2xlarge":      8,
	"r5b.4xlarge":      16,
	"r5b.8xlarge":      32,
	"r5b.large":        2,
	"r5b.metal":        96,
	"r5b.xlarge":       4,
	"r5d.12xlarge":     48,
	"r5d.16xlarge":     64,
	"r5d.24xlarge":     96,
	"r5d.2xlarge":      8,
	"r5d.4xlarge":      16,
	"r5d.8xlarge":      32,
	"r5d.large":        2,
	"r5d.metal":        96,
	"r5d.xlarge":       4,
	"r5dn.12xlarge":    48,
	"r5dn.16xlarge":    64,
	"r5dn.24xlarge":    96,
	"r5dn.2xlarge":     8,
	"r5dn.4xlarge":     16,
	"r5dn.8xlarge":     32,
	"r5dn.large":       2,
	"r5dn.metal":       96,
	"r5dn.xlarge":      4,
	"r5n.12xlarge":     48,
	"r5n.16xlarge":     64,
	"r5n.24xlarge":     96,
	"r5n.2xlarge":      8,
	"r5n.4xlarge":      16,
	"r5n.8xlarge":      32,
	"r5n.large":        2,
	"r5n.metal":        96,
	"r5n.xlarge":       4,
	"r6g.12xlarge":     48,
	"r6g.16xlarge":     64,
	"r6g.2xlarge":      8,
	"r6g.4xlarge":      16,
	"r6g.8xlarge":      32,
	"r6g.large":        2,
	"r6g.medium":       1,
	"r6g.metal":        64,
	"r6g.xlarge":       4,
	"r6gd.12xlarge":    48,
	"r6gd.16xlarge":    64,
	"r6gd.2xlarge":     8,
	"r6gd.4xlarge":     16,
	"r6gd.8xlarge":     32,
	"r6gd.large":       2,
	"r6gd.medium":      1,
	"r6gd.metal":       64,
	"r6gd.xlarge":      4,
	"r6i.12xlarge":     48,
	"r6i.16xlarge":     64,
	"r6i.24xlarge":     96,
	"r6i.2xlarge":      8,
	"r6i.32xlarge":     128,
	"r6i.4xlarge":      16,
	"r6i.8xlarge":      32,
	"r6i.large":        2,
	"r6i.metal":        128,
	"r6i.xlarge":       4,
	"t2.2xlarge":       8,
	"t2.large":         2,
	"t2.medium":        2,
	"t2.micro":         1,
	"t2.nano":          1,
	"t2.small":         1,
	"t2.xlarge":        4,
	"t3.2xlarge":       8,
	"t3.large":         2,
	"t3.medium":        2,
	"t3.micro":         2,
	"t3.nano":          2,
	"t3.small":         2,
	"t3.xlarge":        4,
	"t3a.2xlarge":      8,
	"t3a.large":        2,
	"t3a.medium":       2,
	"t3a.micro":        2,
	"t3a.nano":         2,
	"t3a.small":        2,
	"t3a.xlarge":       4,
	"t4g.2xlarge":      8,
	"t4g.large":        2,
	"t4g.medium":       2,
	"t4g.micro":        2,
	"t4g.nano":         2,
	"t4g.small":        2,
	"t4g.xlarge":       4,
	"u-6tb1.112xlarge": 448,
	"u-6tb1.56xlarge":  224,
	"x1.16xlarge":      64,
	"x1.32xlarge":      128,
	"x1e.16xlarge":     64,
	"x1e.2xlarge":      8,
	"x1e.32xlarge":     128,
	"x1e.4xlarge":      16,
	"x1e.8xlarge":      32,
	"x1e.xlarge":       4,
	"x2gd.12xlarge":    48,
	"x2gd.16xlarge":    64,
	"x2gd.2xlarge":     8,
	"x2gd.4xlarge":     16,
	"x2gd.8xlarge":     32,
	"x2gd.large":       2,
	"x2gd.medium":      1,
	"x2gd.metal":       64,
	"x2gd.xlarge":      4,
	"z1d.12xlarge":     48,
	"z1d.2xlarge":      8,
	"z1d.3xlarge":      12,
	"z1d.6xlarge":      24,
	"z1d.large":        2,
	"z1d.metal":        48,
	"z1d.xlarge":       4,
}

var reservedTermsMapping = map[string]string{
	"1_year": "1yr",
	"3_year": "3yr",
}

var reservedPaymentOptionMapping = map[string]string{
	"no_upfront":      "No Upfront",
	"partial_upfront": "Partial Upfront",
	"all_upfront":     "All Upfront",
}

// There's differences within the pricing API, the Values have no space.
var reservedHostPaymentOptionMapping = map[string]string{
	"no_upfront":      "NoUpfront",
	"partial_upfront": "PartialUpfront",
	"all_upfront":     "AllUpfront",
}

var elasticacheReservedNodeCacheLegacyOfferings = map[string]string{
	"heavy_utilization":  "Heavy Utilization",
	"medium_utilization": "Medium Utilization",
	"light_utilization":  "Light Utilization",
}

var elasticacheReservedNodeLegacyTypes = []string{"t2", "m3", "m4", "r3", "r4"}

type rdsReservationResolver struct {
	term          string
	paymentOption string
}

// PriceFilter implementation for rdsReservationResolver
// Allowed values for ReservedInstanceTerm: ["1_year", "3_year"]
// Allowed values for ReservedInstancePaymentOption: ["all_upfront", "partial_upfront", "no_upfront"]
// Corner case: When ReservedInstanceTerm is 3_year the only allowed ReservedInstancePaymentOption are ["all_upfront", "partial_upfront"]
func (r rdsReservationResolver) PriceFilter() (*schema.PriceFilter, error) {
	purchaseOptionLabel := "reserved"
	def := &schema.PriceFilter{
		PurchaseOption: strPtr(purchaseOptionLabel),
	}
	termLength := reservedTermsMapping[r.term]
	purchaseOption := reservedPaymentOptionMapping[r.paymentOption]
	validTerms := sliceOfKeysFromMap(reservedTermsMapping)
	if !stringInSlice(validTerms, r.term) {
		return def, fmt.Errorf("Invalid reserved_instance_term, ignoring reserved options. Expected: %s. Got: %s", strings.Join(validTerms, ", "), r.term)
	}
	validOptions := sliceOfKeysFromMap(reservedPaymentOptionMapping)
	if r.term == "3_year" {
		validOptions = []string{"partial_upfront", "all_upfront"}
	}
	if !stringInSlice(validOptions, r.paymentOption) {
		return def, fmt.Errorf("Invalid reserved_instance_payment_option, ignoring reserved options. Expected: %s. Got: %s", strings.Join(validOptions, ", "), r.paymentOption)
	}
	return &schema.PriceFilter{
		PurchaseOption:     strPtr(purchaseOptionLabel),
		StartUsageAmount:   strPtr("0"),
		TermLength:         strPtr(termLength),
		TermPurchaseOption: strPtr(purchaseOption),
	}, nil
}

func isAWSChina(region string) bool {
	return strings.HasPrefix(region, "cn-")
}
