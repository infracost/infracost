package resources

import (
	"reflect"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

// Dummy variables for type checking
var intPtr *int64
var floatPtr *float64
var strPtr *string
var strType = reflect.TypeFor[string]()
var float64Type = reflect.TypeFor[float64]()

func PopulateArgsWithUsage(args any, u *schema.UsageData) {
	if u == nil {
		// nothing to do
		return
	}

	ps := reflect.ValueOf(args)
	s := ps.Elem()
	if s.Kind() != reflect.Struct {
		return
	}

	// Iterate over all fields of the args struct and
	// try to set the values from the usage if it exist.
	for i := 0; i < s.NumField(); i++ {
		// Get the i'th field of the struct
		f := s.Field(i)
		if !f.IsValid() {
			continue
		}
		if !f.CanSet() {
			continue
		}
		// Get the infracost tag data.
		infracostTagSplitted := strings.Split(s.Type().Field(i).Tag.Get("infracost_usage"), ",")
		if len(infracostTagSplitted) < 1 {
			continue
		}
		// Key name for the usage file
		usageKey := infracostTagSplitted[0]
		if usageKey == "" {
			continue
		}

		// Check whether a value for this arg was specified in the usage data.
		if u.Get(usageKey).Exists() {
			// Set the value of the arg to the value specified in the usage data.
			if f.Type() == reflect.TypeFor[*float64]() {
				f.Set(reflect.ValueOf(u.GetFloat(usageKey)))
				continue
			}

			if f.Type() == reflect.TypeFor[*int64]() {
				f.Set(reflect.ValueOf(u.GetInt(usageKey)))
				continue
			}

			if f.Type() == reflect.TypeFor[*string]() {
				f.Set(reflect.ValueOf(u.GetString(usageKey)))
				continue
			}

			if f.Type().Elem().Kind() == reflect.Struct {
				if f.IsNil() {
					f.Set(reflect.New(f.Type().Elem()))
				}

				PopulateArgsWithUsage(f.Interface(), &schema.UsageData{
					Address:    usageKey,
					Attributes: u.Get(usageKey).Map(),
				})

				continue
			}

			if f.Type() == reflect.MapOf(strType, float64Type) {
				m := make(map[string]float64)
				for k, v := range u.Get(usageKey).Map() {
					m[k] = v.Float()
				}

				f.Set(reflect.ValueOf(m))

				continue
			}

			logging.Logger.Error().Msgf("Unsupported field { UsageKey: %s, Type: %v }", usageKey, f.Type())
		}
	}
}
