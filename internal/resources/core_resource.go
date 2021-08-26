package resources

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/infracost/infracost/internal/schema"
)

type CoreResource interface {
	PopulateUsage(u *schema.UsageData)
}

func PopulateArgsWithUsage(args interface{}, u *schema.UsageData) {
	ps := reflect.ValueOf(args)
	s := ps.Elem()
	if s.Kind() != reflect.Struct {
		return
	}

	// Dummy variables for type checking
	var floatPtr *float64

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
		fmt.Println(usageKey)

		// The arg is a pointer (*float64, *string, ...)
		if f.Kind() == reflect.Ptr {
			// It's a *float64
			if f.Type() == reflect.TypeOf(floatPtr) {

				// Check whether a value for this arg was specified in the usage data.
				if u != nil && u.Get(usageKey).Exists() {
					// Set the value of the arg to the value specified in the usage data.
					f.Set(reflect.ValueOf(u.GetFloat(usageKey)))
				}
			}
			// TODO: Add support for other kinds (*string, *int, *int64).
		}
	}
}
