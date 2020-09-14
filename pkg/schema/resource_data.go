package schema

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

type ResourceData struct {
	Type          string
	ProviderName  string
	Address       string
	rawValues     gjson.Result
	referencesMap map[string][]*ResourceData
}

func NewResourceData(resourceType string, providerName string, address string, rawValues gjson.Result) *ResourceData {
	return &ResourceData{
		Type:          resourceType,
		ProviderName:  providerName,
		Address:       address,
		rawValues:     rawValues,
		referencesMap: make(map[string][]*ResourceData),
	}
}

func (d *ResourceData) Get(key string) gjson.Result {
	return d.rawValues.Get(key)
}

func (d *ResourceData) References(key string) []*ResourceData {
	return d.referencesMap[key]
}

func (d *ResourceData) AddReference(key string, reference *ResourceData) {
	if _, ok := d.referencesMap[key]; !ok {
		d.referencesMap[key] = make([]*ResourceData, 0)
	}
	d.referencesMap[key] = append(d.referencesMap[key], reference)
}

func (d *ResourceData) Set(key string, value interface{}) {
	d.rawValues = AddRawValue(d.rawValues, key, value)
}

func AddRawValue(r gjson.Result, key string, v interface{}) gjson.Result {
	j := make(map[string]interface{})

	_ = json.Unmarshal([]byte(r.Raw), &j) // TODO: unhandled error
	if j == nil {
		j = make(map[string]interface{})
	}

	j[key] = v

	mj, _ := json.Marshal(j) // TODO: unhandled error

	return gjson.ParseBytes(mj)
}
