package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

type UsageVariableType int

const (
	Int64 UsageVariableType = iota
	String
	Float64
	StringArray
)

// type UsageDataValidatorFuncType = func(value interface{}) error

type UsageSchemaItem struct {
	Key          string
	DefaultValue interface{}
	ValueType    UsageVariableType
	// These aren't used yet and I'm not entirely sure how they fit in, but they were part of the discussion about usage schema.
	// ValidatorFunc UsageDataValidatorFuncType
	// SubUsageData  *UsageSchemaItem
	// Description   string
}

type UsageData struct {
	Address    string
	Attributes map[string]gjson.Result
}

func NewUsageData(address string, attributes map[string]gjson.Result) *UsageData {
	return &UsageData{
		Address:    address,
		Attributes: attributes,
	}
}

func (u *UsageData) Get(key string) gjson.Result {
	if u.Attributes[key].Type != gjson.Null {
		return u.Attributes[key]
	} else if strings.Contains(key, "[") && strings.Contains(key, "]") {
		key = convertArrayKeyToWildcard(key)
	}

	return u.Attributes[key]
}

// GetMap returns a map of gjson.Result. GetMap expects key to be
// the name of a usage file property that contains sub properties. e.g.
//
// 		given a usage file with the following property:
//
//			prop_1:
//				sub_prop_1: 1000
//				sub_prop_2: 3000
//
// 		GetMap with the "prop_1" as key will return a map:
//
//			map[string]gjson.Result{
//				"sub_prop_1": ...,
//				"sub_prop_2": ...,
//		   }
//
// GetMap only support 1 level of nesting. Further levels must be handled another way.
// If key cannot be found in UsageData GetMap returns nil.
func (u *UsageData) GetMap(key string) map[string]gjson.Result {
	start, err := regexp.Compile(`^` + key + `\.`)
	if err != nil {
		return nil
	}

	matching := make(map[string]gjson.Result)

	for s, result := range u.Attributes {
		if start.MatchString(s) {
			matching[start.ReplaceAllString(s, "")] = result
		}
	}

	if len(matching) == 0 {
		return nil
	}

	return matching
}

func (u *UsageData) GetFloat(key string) *float64 {
	if u.Get(key).Type != gjson.Null {
		val := u.Get(key).Float()
		return &val
	}

	return nil
}

func (u *UsageData) GetInt(key string) *int64 {
	if u.Get(key).Type != gjson.Null {
		val := u.Get(key).Int()
		return &val
	}

	return nil
}

func (u *UsageData) GetString(key string) *string {
	if u.Get(key).Type != gjson.Null {
		val := u.Get(key).String()
		return &val
	}

	return nil
}

func (u *UsageData) GetStringArray(key string) *[]string {
	if u.Get(key).Type != gjson.Null {
		gjsonArray := u.Get(key).Array()

		stringArray := make([]string, len(gjsonArray))
		for i, gresult := range gjsonArray {
			stringArray[i] = gresult.String()
		}
		return &stringArray
	}

	return nil
}

func convertArrayKeyToWildcard(key string) string {
	lastOpenBracket := strings.LastIndex(key, "[")
	lastCloseBracket := strings.LastIndex(key, "]")

	return key[:lastOpenBracket+1] + "*" + key[lastCloseBracket:]
}

// CalcEstimationSummary returns a map where a value of true means the attribute key has an actual estimate, false means
// it is using the defaults
func (u *UsageData) CalcEstimationSummary() map[string]bool {
	estimationMap := make(map[string]bool)
	for k, v := range u.Attributes {
		// figure out if the attribute has estimated value or if it is just using the defaults
		hasEstimate := false
		switch v.Type {
		case gjson.Number:
			hasEstimate = v.Num > 0
		case gjson.String:
			hasEstimate = v.Str != ""
		}
		estimationMap[k] = hasEstimate
	}
	return estimationMap
}

func NewUsageMap(m map[string]interface{}) map[string]*UsageData {
	usageMap := make(map[string]*UsageData)

	for addr, v := range m {
		usageMap[addr] = NewUsageData(
			addr,
			ParseAttributes(v),
		)
	}

	return usageMap
}

func NewEmptyUsageMap() map[string]*UsageData {
	return map[string]*UsageData{}
}

func ParseAttributes(i interface{}) map[string]gjson.Result {
	a := make(map[string]gjson.Result)
	for k, v := range flatten(i) {
		j, _ := json.Marshal(v)
		a[k] = gjson.ParseBytes(j)
	}

	return a
}

func flatten(i interface{}) map[string]interface{} {
	keys := make([]string, 0)
	result := make(map[string]interface{})
	flattenHelper(i, keys, result)
	return result
}

func flattenHelper(i interface{}, keys []string, result map[string]interface{}) {
	switch v := i.(type) {
	case map[string]interface{}:
		for k, v := range i.(map[string]interface{}) {
			flattenHelper(v, append(keys, k), result)
		}
	case map[interface{}]interface{}:
		for k, v := range i.(map[interface{}]interface{}) {
			flattenHelper(v, append(keys, fmt.Sprintf("%v", k)), result)
		}
	default:
		k := strings.Join(keys, ".")
		result[k] = v
	}
}
