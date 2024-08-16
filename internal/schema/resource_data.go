package schema

import (
	"encoding/json"
	"strings"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/tidwall/gjson"
)

type ResourceData struct {
	Type                                    string
	ProviderName                            string
	Address                                 string
	Tags                                    *map[string]string
	DefaultTags                             *map[string]string
	ProviderSupportsDefaultTags             bool
	ProviderLink                            string
	TagPropagation                          *TagPropagation
	RawValues                               gjson.Result
	ReferencesMap                           map[string][]*ResourceData
	CFResource                              cloudformation.Resource
	UsageData                               *UsageData
	Metadata                                map[string]gjson.Result
	MissingVarsCausingUnknownTagKeys        []string
	MissingVarsCausingUnknownDefaultTagKeys []string
	// Region is the region of the resource. When building a resource callers should
	// use this value instead of the deprecated d.Get("region").String() or
	// lookupRegion method.
	Region string
}

type TagPropagation struct {
	To                    string             // a human-readable name of the type being propagated to - will not always have a tf type
	From                  *string            // e.g. SERVICE, TASK_DEFINITION
	Tags                  *map[string]string // tags that were propagated from the above resource, if any
	Attribute             string             // the attribute that can be used to configured propagation, e.g. propagate_tags
	HasRequiredAttributes bool               // whether the resource has the required attributes to warrant propagating tags
}

func NewResourceData(resourceType string, providerName string, address string, tags *map[string]string, rawValues gjson.Result) *ResourceData {
	return &ResourceData{
		Type:          resourceType,
		ProviderName:  providerName,
		Address:       address,
		Tags:          tags,
		RawValues:     rawValues,
		ReferencesMap: make(map[string][]*ResourceData),
		CFResource:    nil,
	}
}

func NewCFResourceData(resourceType string, providerName string, address string, tags *map[string]string, cfResource cloudformation.Resource) *ResourceData {
	return &ResourceData{
		Type:          resourceType,
		ProviderName:  providerName,
		Address:       address,
		Tags:          tags,
		RawValues:     gjson.Result{},
		ReferencesMap: make(map[string][]*ResourceData),
		CFResource:    cfResource,
	}
}

func (d *ResourceData) Get(key string) gjson.Result {
	return gjson.Parse(strings.Clone(d.RawValues.Get(key).Raw))
}

// GetStringOrDefault returns the value of key within ResourceData as a string.
// If the retrieved value is not set GetStringOrDefault will return def.
func (d *ResourceData) GetStringOrDefault(key, def string) string {
	if !d.IsEmpty(key) {
		return strings.Clone(d.RawValues.Get(key).String())
	}

	return def
}

// GetInt64OrDefault returns the value of key within ResourceData as an int64.
// If the retrieved value is not set GetInt64OrDefault will return def.
func (d *ResourceData) GetInt64OrDefault(key string, def int64) int64 {
	if !d.IsEmpty(key) {
		return d.RawValues.Get(key).Int()
	}

	return def
}

// GetFloat64OrDefault returns the value of key within ResourceData as a float64.
// If the retrieved value is not set GetFloat64OrDefault will return def.
func (d *ResourceData) GetFloat64OrDefault(key string, def float64) float64 {
	if !d.IsEmpty(key) {
		return d.RawValues.Get(key).Float()
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

func (d *ResourceData) References(keys ...string) []*ResourceData {
	var data []*ResourceData

	for _, key := range keys {
		data = append(data, d.ReferencesMap[key]...)
	}

	return data
}

func (d *ResourceData) AddReference(k string, reference *ResourceData, reverseRefAttrs []string) {
	// Perf/memory leak: Copy gjson string slices that may be returned so we don't prevent
	// the entire underlying parsed json from being garbage collected.
	key := strings.Clone(k)
	if _, ok := d.ReferencesMap[key]; !ok {
		d.ReferencesMap[key] = make([]*ResourceData, 0)
	}
	d.ReferencesMap[key] = append(d.ReferencesMap[key], reference)

	// add any reverse references
	reverseRefKey := d.Type + "." + key
	for _, attr := range reverseRefAttrs {
		if attr == reverseRefKey {
			if _, ok := reference.ReferencesMap[reverseRefKey]; !ok {
				reference.ReferencesMap[reverseRefKey] = make([]*ResourceData, 0)
			}
			reference.ReferencesMap[reverseRefKey] = append(reference.ReferencesMap[reverseRefKey], d)
		}
	}
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
