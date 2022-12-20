package google

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

var (
	vendorName = strPtr("gcp")
	underscore = regexp.MustCompile(`_`)
)

func strPtr(s string) *string {
	return &s
}

// nolint:deadcode,unused
func regexPtr(regex string) *string {
	return strPtr(fmt.Sprintf("/%s/i", regex))
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

// nolint:deadcode,unused
func floatPtrToDecimalPtr(f *float64) *decimal.Decimal {
	if f == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromFloat(*f))
}

// nolint:deadcode,unused
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// RegionsUsage is a reusable type that represents a usage cost map.
// This can be used in resources that define a usage parameter that's changed on a per-region basis, e.g:
//
// monthly_data_processed_gb:
//
//	asia_northeast1: 188
//	asia_east2: 78
//
// can be handled by adding a usage cost property to your resource like so:
//
//	type MyResource struct {
//	   ...
//	   MonthlyDataProcessedGB *RegionsUsage `infracost_usage:"monthly_processed_gb"`
//	}
type RegionsUsage struct {
	AsiaEast1              *float64 `infracost_usage:"asia_east1"`
	AsiaEast2              *float64 `infracost_usage:"asia_east2"`
	AsiaNortheast1         *float64 `infracost_usage:"asia_northeast1"`
	AsiaNortheast2         *float64 `infracost_usage:"asia_northeast2"`
	AsiaNortheast3         *float64 `infracost_usage:"asia_northeast3"`
	AsiaSouth1             *float64 `infracost_usage:"asia_south1"`
	AsiaSouth2             *float64 `infracost_usage:"asia_south2"`
	AsiaSoutheast1         *float64 `infracost_usage:"asia_southeast1"`
	AsiaSoutheast2         *float64 `infracost_usage:"asia_southeast2"`
	AustraliaSoutheast1    *float64 `infracost_usage:"australia_southeast1"`
	AustraliaSoutheast2    *float64 `infracost_usage:"australia_southeast2"`
	EuropeCentral2         *float64 `infracost_usage:"europe_central2"`
	EuropeNorth1           *float64 `infracost_usage:"europe_north1"`
	EuropeWest1            *float64 `infracost_usage:"europe_west1"`
	EuropeWest2            *float64 `infracost_usage:"europe_west2"`
	EuropeWest3            *float64 `infracost_usage:"europe_west3"`
	EuropeWest4            *float64 `infracost_usage:"europe_west4"`
	EuropeWest6            *float64 `infracost_usage:"europe_west6"`
	NorthAmericaNortheast1 *float64 `infracost_usage:"northamerica_northeast1"`
	NorthAmericaNortheast2 *float64 `infracost_usage:"northamerica_northeast2"`
	SouthAmericaEast1      *float64 `infracost_usage:"southamerica_east1"`
	SouthAmericaWest1      *float64 `infracost_usage:"southamerica_west1"`
	USCentral1             *float64 `infracost_usage:"us_central1"`
	USEast1                *float64 `infracost_usage:"us_east1"`
	USEast4                *float64 `infracost_usage:"us_east4"`
	USWest1                *float64 `infracost_usage:"us_west1"`
	USWest2                *float64 `infracost_usage:"us_west2"`
	USWest3                *float64 `infracost_usage:"us_west3"`
	USWest4                *float64 `infracost_usage:"us_west4"`
}

// RegionUsageSchema is the schema representation of the RegionsUsage type.
// This can be used as a schema.SubResourceUsage to define a structure that's
// commonly used with resources that vary on a per region basis.
var RegionUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_east1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_east2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_northeast1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_northeast2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_northeast3"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_south1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_south2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_southeast1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia_southeast2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "australia_southeast1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "australia_southeast2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_central2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_north1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_west1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_west2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_west3"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_west4"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe_west6"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "northamerica_northeast1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "northamerica_northeast2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "southamerica_east1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "southamerica_west1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_central1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_east1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_east4"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_west1"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_west2"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_west3"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_west4"},
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

func GetFloatFieldValueByUsageTag(tagValue string, s interface{}) float64 {
	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		return 0
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := strings.Split(f.Tag.Get("infracost_usage"), ",")[0] // use split to ignore tag "options" like omitempty, etc.
		if v == tagValue {
			r := reflect.ValueOf(s)
			field := reflect.Indirect(r).FieldByName(f.Name)
			if !field.Elem().IsValid() {
				return 0
			}
			return field.Elem().Float()
		}
	}
	return 0
}
