package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13.
var infracostProviderNames = []string{"infracost", "registry.terraform.io/infracost/infracost"}

func createResource(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "This resource is free",
			}
		}

		res := registryItem.RFunc(d, u)
		if res != nil {
			res.ResourceType = d.Type
			return res
		}
	}

	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		IsSkipped:    true,
		SkipMessage:  "This resource is not currently supported",
	}
}

func parsePlanJSON(j []byte) ([]*schema.Resource, error) {
	resources := make([]*schema.Resource, 0)

	if !gjson.ValidBytes(j) {
		return resources, errors.New("invalid JSON")
	}

	p := gjson.ParseBytes(j)
	providerConf := p.Get("configuration.provider_config")
	planVals := p.Get("planned_values.root_module")
	conf := p.Get("configuration.root_module")
	vars := p.Get("variables")

	resData := parseResourceData(providerConf, planVals, conf, vars)
	parseReferences(resData, conf)
	resUsage := buildUsageResourceDataMap(resData)
	resData = stripInfracostResources(resData)

	for _, r := range resData {
		if res := createResource(r, resUsage[r.Address]); res != nil {
			resources = append(resources, res)
		}
	}

	return resources, nil
}

func parseResourceData(providerConf, planVals gjson.Result, conf gjson.Result, vars gjson.Result) map[string]*schema.ResourceData {
	defaultRegion := parseProviderRegion(providerConf, "aws", vars)
	if defaultRegion == "" {
		defaultRegion = "us-east-1"
	}

	resources := make(map[string]*schema.ResourceData)

	for _, r := range planVals.Get("resources").Array() {
		t := r.Get("type").String()
		provider := r.Get("provider_name").String()
		addr := r.Get("address").String()
		v := r.Get("values")

		region := defaultRegion

		// Override the region with the provider alias's region if exists
		resConf := getConfJSON(conf, addr)

		providerKey := parseProviderKey(resConf)
		if providerKey != "aws" && providerKey != "" {
			provRegion := parseProviderRegion(providerConf, providerKey, vars)
			// Note: if the provider is passed to a module using a different alias
			// then there's no way to detect this so we just have to fallback to
			// the default provider
			if provRegion != "" {
				region = provRegion
			}
		}

		// Override the region with the region from the arn if exists
		if v.Get("arn").Exists() {
			region = strings.Split(v.Get("arn").String(), ":")[3]
		}

		v = schema.AddRawValue(v, "region", region)

		resources[addr] = schema.NewResourceData(t, provider, addr, v)
	}

	// Recursively add any resources for child modules
	for _, m := range planVals.Get("child_modules").Array() {
		for addr, d := range parseResourceData(providerConf, m, conf, vars) {
			resources[addr] = d
		}
	}

	return resources
}

func parseProviderKey(resConf gjson.Result) string {
	v := resConf.Get("provider_config_key").String()
	p := strings.Split(v, ":")

	return p[len(p)-1]
}

func parseProviderRegion(providerConfig gjson.Result, providerKey string, vars gjson.Result) string {
	// Try to get constant value
	region := providerConfig.Get(fmt.Sprintf("%s.expressions.region.constant_value", gjsonEscape(providerKey))).String()
	if region == "" {
		// Try to get reference
		refName := providerConfig.Get(fmt.Sprintf("%s.expressions.region.references.0", gjsonEscape(providerKey))).String()
		splitRef := strings.Split(refName, ".")
		if splitRef[0] == "var" {
			// Get the region from variables
			varName := strings.Join(splitRef[1:], ".")
			varContent := vars.Get(fmt.Sprintf("%s.value", varName))
			if !varContent.IsObject() && !varContent.IsArray() {
				region = varContent.String()
			}
		}
	}
	return region
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

	for addr, d := range resData {
		if !isInfracostResource(d) {
			n[addr] = d
		}
	}

	return n
}

func parseReferences(resData map[string]*schema.ResourceData, conf gjson.Result) {
	for addr, res := range resData {
		resConf := getConfJSON(conf, addr)

		var refsMap = make(map[string][]string)

		for attr, j := range resConf.Get("expressions").Map() {
			getReferences(res, attr, j, &refsMap)
		}

		for attr, refs := range refsMap {
			for _, ref := range refs {
				if ref == "count.index" || strings.HasPrefix(ref, "var.") {
					continue
				}

				var refData *schema.ResourceData

				m := addressModulePart(addr)
				refAddr := fmt.Sprintf("%s%s", m, ref)

				// see if there's a resource that's an exact match on the address
				refData, ok := resData[refAddr]

				// if there's a count ref value then try with the array index of the count ref
				if !ok && containsString(refs, "count.index") {
					a := fmt.Sprintf("%s[%d]", refAddr, addressCountIndex(addr))
					refData, ok = resData[a]

					if ok {
						log.Debugf("reference specifies a count: using resource %s for %s.%s", a, addr, attr)
					}
				}

				// if still not found, see if there's a matching resource with an [0] array part
				if !ok {
					a := fmt.Sprintf("%s[0]", refAddr)
					refData, ok = resData[a]

					if ok {
						log.Debugf("reference does not specify a count: using resource %s for for %s.%s", a, addr, attr)
					}
				}

				if ok {
					res.AddReference(attr, refData)
				}
			}
		}
	}
}

func getReferences(resData *schema.ResourceData, attr string, j gjson.Result, refs *map[string][]string) {
	if j.Get("references").Exists() {
		for _, ref := range j.Get("references").Array() {
			if _, ok := (*refs)[attr]; !ok {
				(*refs)[attr] = make([]string, 0, 1)
			}

			(*refs)[attr] = append((*refs)[attr], ref.String())
		}
	} else if j.IsArray() {
		for i, attributeJSONItem := range j.Array() {
			getReferences(resData, fmt.Sprintf("%s.%d", attr, i), attributeJSONItem, refs)
		}
	} else if j.Type.String() == "JSON" {
		j.ForEach(func(childAttribute gjson.Result, childAttributeJSON gjson.Result) bool {
			getReferences(resData, fmt.Sprintf("%s.%s", attr, childAttribute), childAttributeJSON, refs)

			return true
		})
	}
}

func getConfJSON(conf gjson.Result, addr string) gjson.Result {
	c := getModuleConfJSON(conf, getModuleNames(addr))

	return c.Get(fmt.Sprintf(`resources.#(address="%s")`, removeAddressArrayPart(addressResourcePart(addr))))
}

func getModuleConfJSON(conf gjson.Result, names []string) gjson.Result {
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

// addressResourcePart parses a resource addr and returns resource suffix (without the module prefix).
// For example: `module.name1.module.name2.resource` will return `name2.resource`.
func addressResourcePart(addr string) string {
	p := strings.Split(addr, ".")

	if len(p) >= 3 && p[len(p)-3] == "data" {
		return strings.Join(p[len(p)-3:], ".")
	}

	return strings.Join(p[len(p)-2:], ".")
}

// addressModulePart parses a resource addr and returns module prefix.
// For example: `module.name1.module.name2.resource` will return `module.name1.module.name2.`.
func addressModulePart(addr string) string {
	ap := strings.Split(addr, ".")

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

func getModuleNames(addr string) []string {
	r := regexp.MustCompile(`module\.([^\.\[]*)`)
	matches := r.FindAllStringSubmatch(addressModulePart(addr), -1)

	if matches == nil {
		return []string{}
	}

	n := make([]string, 0, len(matches))
	for _, m := range matches {
		n = append(n, m[1])
	}

	return n
}

func addressCountIndex(addr string) int {
	r := regexp.MustCompile(`\[(\d+)\]`)
	m := r.FindStringSubmatch(addr)

	if len(m) > 0 {
		i, _ := strconv.Atoi(m[1]) // TODO: unhandled error

		return i
	}

	return -1
}

func removeAddressArrayPart(addr string) string {
	r := regexp.MustCompile(`([^\[]+)`)
	m := r.FindStringSubmatch(addressResourcePart(addr))

	return m[1]
}

func containsString(a []string, s string) bool {
	for _, i := range a {
		if i == s {
			return true
		}
	}

	return false
}

func gjsonEscape(s string) string {
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "*", `\*`)
	s = strings.ReplaceAll(s, "?", `\?`)

	return s
}
