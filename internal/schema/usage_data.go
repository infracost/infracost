package schema

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"regexp"
	"sort"
	"strings"

	addressParser "github.com/hashicorp/go-terraform-address"
	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
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

// Copy returns a clone of UsageData u.
func (u *UsageData) Copy() *UsageData {
	if u == nil {
		return nil
	}

	c := &UsageData{
		Address:    u.Address,
		Attributes: map[string]gjson.Result{},
	}

	maps.Copy(c.Attributes, u.Attributes)

	return c
}

// Merge returns a new UsageData which is the result of adding all keys from other that do not already exists in the usage data
func (u *UsageData) Merge(other *UsageData) *UsageData {
	if u == nil {
		if other != nil {
			return other.Copy()
		}

		return nil // both are nil
	}

	newU := &UsageData{
		Address:    u.Address,
		Attributes: make(map[string]gjson.Result, len(u.Attributes)),
	}

	maps.Copy(newU.Attributes, u.Attributes)

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
	if u == nil {
		return gjson.Result{}
	}

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

// UsageMap is a map of address to UsageData built from a usage file.
// UsageMap is a standalone type so that we can do more involved matching functionality.
type UsageMap struct {
	data      map[string]*UsageData
	wildcards wildcards
}

// NewUsageMapFromInterface returns an initialised UsageMap from interface map.
func NewUsageMapFromInterface(m map[string]any) UsageMap {
	data := make(map[string]*UsageData)

	for addr, v := range m {
		data[addr] = NewUsageData(
			addr,
			ParseAttributes(v),
		)
	}

	return NewUsageMap(data)
}

// NewUsageMap initialises a Usage map with the provided usage key data.
// It builds a set of wildcard keys if any are found and sorts them ready for searching
// by Attribute name at a later point.
func NewUsageMap(data map[string]*UsageData) UsageMap {
	var keys wildcards

	for key := range data {
		if strings.Contains(key, "*") {
			keys = append(keys, wildcard{
				raw:    key,
				regexp: usageKeyToRegexp(key),
			})
		}
	}

	sort.Sort(keys)

	return UsageMap{data: data, wildcards: keys}
}

// Data returns the entire map of usage data stored.
func (usage UsageMap) Data() map[string]*UsageData {
	return usage.data
}

// Get returns UsageData for a given resource address, this can be a combined/merged UsageData from multiple keys.
// Usage data is merged adhering to the following hierarchy:
//
//  1. Resource type defaults - e.g. aws_lambda:
//  2. Wildcard specified data - e.g. aws_lambda.my_lambda[*]
//  3. Exact resource data - e.g. aws_lambda.my_lambda["foo"]
//
// Duplicate keys specified between levels are always overwritten by keys specified at a lower level, e.g:
//
//	aws_lambda.my_lambda[*]:
//		monthly_requests: 700000000
//		request_duration_ms: 750
//	aws_lambda.my_lambda["foo"]:
//		request_duration_ms: 100 << this overwrites the 750 value given in the wildcard usage
//
// If no usage key is found, Get will return nil.
func (usage UsageMap) Get(address string) *UsageData {
	var data *UsageData

	parsedAddress, err := addressParser.NewAddress(address)
	if err == nil {
		val, ok := usage.data[parsedAddress.ResourceSpec.Type]
		if ok {
			data = val.Copy()
		}
	}

	for _, key := range usage.wildcards {
		d := usage.data[key.raw]

		if key.regexp.MatchString(address) {
			if data != nil {
				mergeUsage(data, d)
			} else {
				data = d.Copy()
			}

			break
		}
	}

	if ud := usage.data[address]; ud != nil {
		if data != nil {
			mergeUsage(data, ud)
		} else {
			data = ud.Copy()
		}
	}

	return data
}

var wildCardRegxp = regexp.MustCompile(`\[.*?]`)

// wildcard contains information about a wildcard specified usage key.
type wildcard struct {
	raw    string
	regexp *regexp.Regexp
}

// wildcards implements the Sort.Interface to sort a slice of usage keys with more specific keys first.
// This means a list with the following:
//
//   - module.mod[*].resource.test[*]
//   - module.mod["foo"].resource.test["baz]
//   - module.mod["foo].resource.test[*]
//
// will be reordered to the following:
//
//   - module.mod["foo"].resource.test["baz]
//   - module.mod["foo].resource.test[*]
//   - module.mod[*].resource.test[*]
type wildcards []wildcard

func (w wildcards) Len() int {
	return len(w)
}

func (w wildcards) Less(i, j int) bool {
	a := w[i].raw
	b := w[j].raw

	aSplit := wildCardRegxp.Split(a, -1)
	bSplit := wildCardRegxp.Split(b, -1)

	// if these aren't the same resource key then return false.
	aJoined := strings.Join(aSplit, "")
	bJoined := strings.Join(bSplit, "")
	if aJoined != bJoined {
		return aJoined < bJoined
	}

	aKeys := wildCardRegxp.FindAllString(a, -1)
	bKeys := wildCardRegxp.FindAllString(b, -1)

	for index, key := range aKeys {
		if bKeys[index] == key {
			continue
		}

		if key == "[*]" {
			return false
		}
	}

	return true
}

func (w wildcards) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func usageKeyToRegexp(pattern string) *regexp.Regexp {
	var result strings.Builder
	for i, literal := range strings.Split(pattern, "*") {
		if i > 0 {
			result.WriteString(".*")
		}

		// QuoteMeta escapes all regular expression metacharacters so that we don't match things like [ or .
		result.WriteString(regexp.QuoteMeta(literal))
	}

	return regexp.MustCompile("^" + result.String() + "$")
}

func mergeUsage(dst *UsageData, src *UsageData) {
	for key, srcAttr := range src.Attributes {
		if _, ok := dst.Attributes[key]; !ok {
			dst.Attributes[key] = srcAttr
			continue
		}

		switch srcAttr.Type {
		case gjson.String, gjson.Number, gjson.False, gjson.True, gjson.Null:
			// Should be safe to override
			dst.Attributes[key] = srcAttr
		case gjson.JSON:
			var err error
			var destJson map[string]any
			var srcJson map[string]any
			err = json.Unmarshal([]byte(dst.Attributes[key].Raw), &destJson)
			if err != nil {
				logging.Logger.Err(err).Msgf("failed to merge UsageData attributes, could not unmarshal dst attribute key: %q", key)
				continue
			}
			err = json.Unmarshal([]byte(srcAttr.Raw), &srcJson)
			if err != nil {
				logging.Logger.Err(err).Msgf("failed to merge UsageData attributes, could not unmarshal src attribute key: %q", key)
				continue
			}
			err = mergo.Map(&destJson, srcJson)
			if err != nil {
				logging.Logger.Err(err).Msgf("failed to merge UsageData attributes, could not merge attribute key: %q", key)
				continue
			}
			src, err := json.Marshal(destJson)
			if err != nil {
				logging.Logger.Err(err).Msgf("failed to merge UsageData attributes, could not marshal attribute key: %q", key)
				continue
			}

			dst.Attributes[key] = gjson.Parse(string(src))
		}
	}

	dst.Address = src.Address
}

func ParseAttributes(i any) map[string]gjson.Result {
	a := make(map[string]gjson.Result)

	switch attrs := i.(type) {
	case map[string]any:
		for k, v := range attrs {
			// we use the jsoniter lib here as the std lib json
			// cannot handle marshalling map[interface{}]interface{}
			j, _ := jsoniter.Marshal(v)
			a[k] = gjson.ParseBytes(j)
		}
	case map[any]any:
		for k, v := range attrs {
			j, _ := jsoniter.Marshal(v)
			a[fmt.Sprintf("%s", k)] = gjson.ParseBytes(j)
		}
	}

	return a
}
