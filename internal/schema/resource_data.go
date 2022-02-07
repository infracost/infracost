package schema

import (
	"encoding/json"

	"github.com/awslabs/goformation/v4/cloudformation"

	"github.com/tidwall/gjson"
)

type ResourceData struct {
	Type          string
	ProviderName  string
	Address       string
	Tags          map[string]string
	RawValues     gjson.Result
	referencesMap map[string][]*ResourceData
	CFResource    cloudformation.Resource
}

func NewResourceData(resourceType string, providerName string, address string, tags map[string]string, rawValues gjson.Result) *ResourceData {
	return &ResourceData{
		Type:          resourceType,
		ProviderName:  providerName,
		Address:       address,
		Tags:          tags,
		RawValues:     rawValues,
		referencesMap: make(map[string][]*ResourceData),
		CFResource:    nil,
	}
}

func NewCFResourceData(resourceType string, providerName string, address string, tags map[string]string, cfResource cloudformation.Resource) *ResourceData {
	return &ResourceData{
		Type:          resourceType,
		ProviderName:  providerName,
		Address:       address,
		Tags:          tags,
		RawValues:     gjson.Result{},
		referencesMap: make(map[string][]*ResourceData),
		CFResource:    cfResource,
	}
}

func (d *ResourceData) Get(key string) gjson.Result {
	return d.RawValues.Get(key)
}

func (d *ResourceData) GetStringOrDefault(key, def string) string {
	if !d.IsEmpty(key) {
		return d.RawValues.Get(key).String()
	}

	return def
}

func (d *ResourceData) GetInt64OrDefault(key string, def int64) int64 {
	if !d.IsEmpty(key) {
		return d.RawValues.Get(key).Int()
	}

	return def
}

func (d *ResourceData) GetBoolOrDefault(key string, def bool) bool {
	if !d.IsEmpty(key) {
		return d.RawValues.Get(key).Bool()
	}

	return def
}

// Return true if the key doesn't exist, is null, or is an empty string.
// Needed because gjson.Exists returns true as long as a key exists, even if it's empty or null.
func (d *ResourceData) IsEmpty(key string) bool {
	g := d.RawValues.Get(key)
	return g.Type == gjson.Null || len(g.Raw) == 0 || g.Raw == "\"\"" || emptyObjectOrArray(g)
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
	d.RawValues = AddRawValue(d.RawValues, key, value)
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
