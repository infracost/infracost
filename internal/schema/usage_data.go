package schema

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
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

// Merge returns a new UsageData which is the result of adding all keys from other that do not already exists in the usage data
func (u *UsageData) Merge(other *UsageData) *UsageData {
	if u == nil {
		if other != nil {
			return other.Merge(u) // this will return a new copy of other
		}
		return nil // both are nil
	}

	newU := &UsageData{
		Address:    u.Address,
		Attributes: make(map[string]gjson.Result, len(u.Attributes)),
	}

	for k, v := range u.Attributes {
		newU.Attributes[k] = v
	}

	if other != nil {
		for k, v := range other.Attributes {
			if _, ok := newU.Attributes[k]; !ok {
				newU.Attributes[k] = v
			}
		}
	}

	return newU
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
		if val == 0 {
			fVal := u.Get(key).Float()
			val = int64(math.Floor(fVal))
		}
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

func MergeAttributes(dst *UsageData, src *UsageData) {
	for key, srcAttr := range src.Attributes {
		if _, has := dst.Attributes[key]; has {
			switch srcAttr.Type {
			case gjson.Null:
				fallthrough
			case gjson.True:
				fallthrough
			case gjson.False:
				fallthrough
			case gjson.Number:
				fallthrough
			case gjson.String:
				// Should be safe to override
				dst.Attributes[key] = srcAttr
			case gjson.JSON:
				var err error
				var destJson map[string]interface{}
				var srcJson map[string]interface{}
				err = json.Unmarshal([]byte(dst.Attributes[key].Raw), &destJson)
				if err != nil {
					log.Errorf("Error merging attribute '%s': %v", key, err)
					break
				}
				err = json.Unmarshal([]byte(srcAttr.Raw), &srcJson)
				if err != nil {
					log.Errorf("Error merging attribute '%s': %v", key, err)
					break
				}
				err = mergo.Map(&destJson, srcJson)
				if err != nil {
					log.Errorf("Error merging attribute '%s': %v", key, err)
					break
				}
				src, err := json.Marshal(destJson)
				if err != nil {
					log.Errorf("Error merging attribute '%s': %v", key, err)
					break
				}
				dst.Attributes[key] = gjson.Parse(string(src))
			}
		} else {
			dst.Attributes[key] = srcAttr
		}
	}
}
