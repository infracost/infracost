package infracost

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

type CoreResource interface {
	PopulateArgs(u *schema.UsageData)
}

// PopulateDefaultArgsAndUsage tries to do two tasks.
// 1. Try to fill default values of args fields based on provided tags.
// 2. If the field is present in usage data, set the relevant field.
func PopulateDefaultArgsAndUsage(args interface{}, u *schema.UsageData) {
	ps := reflect.ValueOf(args)
	s := ps.Elem()
	if s.Kind() != reflect.Struct {
		return
	}

	// Dummy variables for type checking
	var floatPtr *float64

	// Iterate over all fields of the args struct and
	// try to set the default values.
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
		if len(infracostTagSplitted) < 2 {
			// The infracost tag is invalid.
			// TODO: Log a warning
			continue
		}
		// Key name for the usage file
		usageKey := infracostTagSplitted[0]
		// Default value for the arg
		defaultUsageStr := infracostTagSplitted[1]

		// The arg is a pointer (*float64, *string, ...)
		if f.Kind() == reflect.Ptr {
			// It's a *float64
			if f.Type() == reflect.TypeOf(floatPtr) {
				// Cast the default value to the right type.
				newValue, err := strconv.ParseFloat(defaultUsageStr, 64)
				if err != nil {
					ui.PrintWarningf("Invalid default value for field %v", s.Type().Field(i).Name)
				} else {
					// Set the default value for the tag.
					f.Set(reflect.ValueOf(&newValue))
				}
				// Check whether a value for this arg was specified in the usage data.
				if u.Get(usageKey).Exists() {
					// Set the value of the arg to the value specified in the usage data.
					f.Set(reflect.ValueOf(u.GetFloat(usageKey)))
				}
			}
			// TODO: Add support for other kinds (*string, *int, *int64).
		}
	}
}
