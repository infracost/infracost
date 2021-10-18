package schema

import (
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

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

// Return true if the key doesn't exist, is null, or is an empty string.
// Needed because gjson.Exists returns true as long as a key exists, even if it's empty or null.
func (u *UsageData) IsEmpty(key string) bool {
	g := u.Get(key)
	return g.Type == gjson.Null || len(g.Raw) == 0 || g.Raw == "\"\"" || emptyObjectOrArray(g)
}

func emptyObjectOrArray(g gjson.Result) bool {
	empty := true
	g.ForEach(func(key, val gjson.Result) bool {
		empty = false
		return false // exit the ForEach after the first value
	})
	return empty
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

	switch attrs := i.(type) {
	case map[string]interface{}:
		for k, v := range attrs {
			// we use the jsoniter lib here as the std lib json
			// cannot handle marshalling map[interface{}]interface{}
			j, _ := jsoniter.Marshal(v)
			a[k] = gjson.ParseBytes(j)
		}
	case map[interface{}]interface{}:
		for k, v := range attrs {
			j, _ := jsoniter.Marshal(v)
			a[fmt.Sprintf("%s", k)] = gjson.ParseBytes(j)
		}
	}

	return a
}
