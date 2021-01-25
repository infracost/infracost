package schema

import (
	"encoding/json"
	"fmt"
	"strings"

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
	return u.Attributes[key]
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
