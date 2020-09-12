package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/tidwall/gjson"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13
var infracostProviderNames = []string{"infracost", "infracost.io/infracost/infracost"}

func createResource(resourceData *schema.ResourceData, usageData *schema.ResourceData) *schema.Resource {
	resourceRegistry := getResourceRegistry()

	if rFunc, ok := (*resourceRegistry)[resourceData.Type]; ok {
		return rFunc(resourceData, usageData)
	}

	return nil
}

func parsePlanJSON(j []byte) []*schema.Resource {
	p := gjson.ParseBytes(j)
	pc := p.Get("configuration.provider_config")
	v := p.Get("planned_values.root_module")
	c := p.Get("configuration.root_module")

	resources := make([]*schema.Resource, 0)

	resData := parseResourceData(p, pc, v)
	parseReferences(resData, c)
	resUsage := buildUsageResourceDataMap(resData)
	resData = stripInfracostResources(resData)

	for _, res := range resData {
		if r := createResource(res, resUsage[res.Address]); r != nil {
			resources = append(resources, r)
		}
	}

	return resources
}

func parseResourceData(plan, provider, values gjson.Result) map[string]*schema.ResourceData {
	defaultRegion := parseAwsRegion(provider)

	resources := make(map[string]*schema.ResourceData)

	for _, r := range values.Get("resources").Array() {
		t := r.Get("type").String()
		n := r.Get("provider_name").String()
		a := r.Get("address").String()
		v := r.Get("values")

		// Override the region with the region from the arn if exists
		region := defaultRegion
		if v.Get("arn").Exists() {
			region = strings.Split(v.Get("arn").String(), ":")[3]
		}
		v = schema.AddRawValue(v, "region", region)

		resources[a] = schema.NewResourceData(t, n, a, v)
	}

	// Recursively add any resources for child modules
	for _, m := range values.Get("child_modules").Array() {
		resources := parseResourceData(plan, provider, m)
		for address, d := range resources {
			resources[address] = d
		}
	}
	return resources
}

func parseAwsRegion(providerConfig gjson.Result) string {
	// Find region from terraform provider config
	awsRegion := providerConfig.Get("aws.expressions.region.constant_value").String()
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	return awsRegion
}

func buildUsageResourceDataMap(resData map[string]*schema.ResourceData) map[string]*schema.ResourceData {
	u := make(map[string]*schema.ResourceData)

	for _, r := range resData {
		if isInfracostResource(r) {
			for _, ref := range r.References("resources") {
				u[ref.Address] = r
			}
		}
	}

	return u
}

func stripInfracostResources(resData map[string]*schema.ResourceData) map[string]*schema.ResourceData {
	n := make(map[string]*schema.ResourceData)

	for a, d := range resData {
		if !isInfracostResource(d) {
			n[a] = d
		}
	}

	return n
}

func parseReferences(resData map[string]*schema.ResourceData, conf gjson.Result) {
	for adrs, res := range resData {
		resConf := getConfigurationJSONForResourceAddress(conf, adrs)

		var refsMap = make(map[string][]string)
		for a, j := range resConf.Get("expressions").Map() {
			getReferences(res, a, j, &refsMap)
		}

		for attr, refs := range refsMap {
			for _, ref := range refs {
				if ref == "count.index" {
					continue
				}

				var refAdrs string
				if containsString(refs, "count.index") {
					refAdrs = fmt.Sprintf("%s%s[%d]", addressModulePart(adrs), ref, addressCountIndex(adrs))
				} else {
					refAdrs = fmt.Sprintf("%s%s", addressModulePart(adrs), ref)
				}

				if refData, ok := resData[refAdrs]; ok {
					res.AddReference(attr, refData)
				}
			}
		}
	}
}

func getReferences(resData *schema.ResourceData, attr string, attrJSON gjson.Result, refs *map[string][]string) {
	if attrJSON.Get("references").Exists() {
		for _, ref := range attrJSON.Get("references").Array() {
			if _, ok := (*refs)[attr]; !ok {
				(*refs)[attr] = make([]string, 0, 1)
			}

			(*refs)[attr] = append((*refs)[attr], ref.String())
		}
	} else if attrJSON.IsArray() {
		for i, attributeJSONItem := range attrJSON.Array() {
			getReferences(resData, fmt.Sprintf("%s.%d", attr, i), attributeJSONItem, refs)
		}
	} else if attrJSON.Type.String() == "JSON" {
		attrJSON.ForEach(func(childAttribute gjson.Result, childAttributeJSON gjson.Result) bool {
			getReferences(resData, fmt.Sprintf("%s.%s", attr, childAttribute), childAttributeJSON, refs)

			return true
		})
	}
}

func getConfigurationJSONForResourceAddress(conf gjson.Result, address string) gjson.Result {
	c := getConfigurationJSONForModulePath(conf, getModuleNames(address))

	return c.Get(fmt.Sprintf(`resources.#(address="%s")`, removeAddressArrayPart(addressResourcePart(address))))
}

func getConfigurationJSONForModulePath(conf gjson.Result, names []string) gjson.Result {
	if len(names) == 0 {
		return conf
	}

	// Build up the gjson search key
	p := make([]string, 0, len(names))
	for _, n := range names {
		p = append(p, fmt.Sprintf("module_calls.%s.module", n))
	}

	return conf.Get(strings.Join(p, "."))
}

func isInfracostResource(res *schema.ResourceData) bool {
	for _, p := range infracostProviderNames {
		if res.ProviderName == p {
			return true
		}
	}

	return false
}

// addressResourcePart parses a resource address and returns resource suffix (without the module prefix).
// For example: `module.name1.module.name2.resource` will return `name2.resource`
func addressResourcePart(address string) string {
	p := strings.Split(address, ".")

	if len(p) >= 3 && p[len(p)-3] == "data" {
		return strings.Join(p[len(p)-3:], ".")
	}

	return strings.Join(p[len(p)-2:], ".")
}

// addressModulePart parses a resource address and returns module prefix.
// For example: `module.name1.module.name2.resource` will return `module.name1.module.name2.`
func addressModulePart(address string) string {
	ap := strings.Split(address, ".")
	var mp []string

	if len(ap) >= 3 && ap[len(ap)-3] == "data" {
		mp = ap[:len(ap)-3]
	} else {
		mp = ap[:len(ap)-2]
	}

	if len(mp) == 0 {
		return ""
	}

	return fmt.Sprintf("%s.", strings.Join(mp, "."))
}

func getModuleNames(address string) []string {
	r := regexp.MustCompile(`module\.([^\[]*)`)
	matches := r.FindAllStringSubmatch(addressModulePart(address), -1)
	if matches == nil {
		return []string{}
	}

	n := make([]string, 0, len(matches))
	for _, match := range matches {
		n = append(n, match[1])
	}

	return n
}

func addressCountIndex(address string) int {
	r := regexp.MustCompile(`\[(\d+)\]`)
	m := r.FindStringSubmatch(address)

	if len(m) > 0 {
		i, _ := strconv.Atoi(m[1]) // TODO: unhandled error

		return i
	}

	return -1
}

func removeAddressArrayPart(address string) string {
	r := regexp.MustCompile(`([^\[]+)`)
	match := r.FindStringSubmatch(addressResourcePart(address))

	return match[1]
}

func containsString(a []string, s string) bool {
	for _, i := range a {
		if i == s {
			return true
		}
	}

	return false
}
