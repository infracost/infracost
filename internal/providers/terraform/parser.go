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

// ARN attribute mapping for resources that don't have a standard 'arn' attribute
var arnAttributeMap = map[string]string{
	"cloudwatch_dashboard":     "dashboard_arn",
	"db_snapshot":              "db_snapshot_arn",
	"db_cluster_snapshot":      "db_cluster_snapshot_arn",
	"neptune_cluster_snapshot": "db_cluster_snapshot_arn",
	"docdb_cluster_snapshot":   "db_cluster_snapshot_arn",
	"dms_certificate":          "certificate_arn",
	"dms_endpoint":             "endpoint_arn",
	"dms_replication_instance": "replication_instance_arn",
	"dms_replication_task":     "replication_task_arn",
}

func createResource(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
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

func parseJSON(j []byte) ([]*schema.Resource, error) {
	resources := make([]*schema.Resource, 0)

	if !gjson.ValidBytes(j) {
		return resources, errors.New("invalid JSON")
	}

	p := gjson.ParseBytes(j)
	providerConf := p.Get("configuration.provider_config")
	conf := p.Get("configuration.root_module")
	vars := p.Get("variables")

	vals := p.Get("planned_values.root_module")
	if !vals.Exists() {
		vals = p.Get("values.root_module")
	}

	resData := parseResourceData(providerConf, vals, conf, vars)
	parseReferences(resData, conf)
	resUsage := buildUsageResourceDataMap(resData)
	resData = stripDataResources(resData)

	for _, r := range resData {
		if res := createResource(r, resUsage[r.Address]); res != nil {
			resources = append(resources, res)
		}
	}

	return resources, nil
}

func parseResourceData(providerConf, planVals gjson.Result, conf gjson.Result, vars gjson.Result) map[string]*schema.ResourceData {
	defaultRegion := "us-east-1"
	defaultProviderRegion := parseProviderRegion(providerConf, "aws", vars)

	resources := make(map[string]*schema.ResourceData)

	for _, r := range planVals.Get("resources").Array() {
		t := r.Get("type").String()
		provider := r.Get("provider_name").String()
		addr := r.Get("address").String()
		v := r.Get("values")

		var region string

		// Try getting the region from the ARN
		arnAttr, ok := arnAttributeMap[t]
		if !ok {
			arnAttr = "arn"
		}

		if v.Get(arnAttr).Exists() {
			region = strings.Split(v.Get(arnAttr).String(), ":")[3]
		}

		// Try getting the region from the provider conf
		if region == "" {
			resConf := getConfJSON(conf, addr)
			providerKey := parseProviderKey(resConf)

			if providerKey == "aws" {
				region = defaultProviderRegion
			} else if providerKey != "" {
				region = parseProviderRegion(providerConf, providerKey, vars)
				// Note: if the provider is passed to a module using a different alias
				// then there's no way to detect this so we just have to fallback to
				// the default provider
			}
		}

		// Fallback to default region
		if region == "" {
			log.Debugf("Falling back to default region (%s) for %s", defaultRegion, addr)
			region = defaultRegion
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

	for _, d := range resData {
		if isInfracostResource(d) {
			for _, ref := range d.References("resources") {
				u[ref.Address] = d
			}
		}
	}

	return u
}

func stripDataResources(resData map[string]*schema.ResourceData) map[string]*schema.ResourceData {
	n := make(map[string]*schema.ResourceData)

	for addr, d := range resData {
		if !strings.HasPrefix(d.Address, "data.") {
			n[addr] = d
		}
	}

	return n
}

func parseReferences(resData map[string]*schema.ResourceData, conf gjson.Result) {
	registryMap := GetResourceRegistryMap()

	// Create a map of id -> resource data so we can lookup references
	idMap := make(map[string]*schema.ResourceData)
	for _, d := range resData {
		idMap[d.Get("id").String()] = d
	}

	for _, d := range resData {
		var refAttrs []string

		if isInfracostResource(d) {
			refAttrs = []string{"resources"}
		} else {
			item, ok := (*registryMap)[d.Type]
			if ok {
				refAttrs = item.ReferenceAttributes
			}
		}

		for _, attr := range refAttrs {
			found := parseConfReferences(resData, conf, d, attr)

			if found {
				continue
			}

			// Get any values for the fields and check if they map to IDs of any resources
			for _, refID := range d.Get(attr).Array() {
				refData, ok := idMap[refID.String()]
				if ok {
					d.AddReference(attr, refData)
				}
			}
		}
	}
}

func parseConfReferences(resData map[string]*schema.ResourceData, conf gjson.Result, d *schema.ResourceData, attr string) bool {
	// Check if there's a reference in the conf
	resConf := getConfJSON(conf, d.Address)
	refResults := resConf.Get("expressions").Get(attr).Get("references").Array()
	refs := make([]string, 0, len(refResults))

	for _, refR := range refResults {
		refs = append(refs, refR.String())
	}

	found := false

	for _, ref := range refs {
		if ref == "count.index" || strings.HasPrefix(ref, "var.") {
			continue
		}

		var refData *schema.ResourceData

		m := addressModulePart(d.Address)
		refAddr := fmt.Sprintf("%s%s", m, ref)

		// see if there's a resource that's an exact match on the address
		refData, ok := resData[refAddr]

		// if there's a count ref value then try with the array index of the count ref
		if !ok && containsString(refs, "count.index") {
			a := fmt.Sprintf("%s[%d]", refAddr, addressCountIndex(d.Address))
			refData, ok = resData[a]

			if ok {
				log.Debugf("reference specifies a count: using resource %s for %s.%s", a, d.Address, attr)
			}
		}

		// if still not found, see if there's a matching resource with an [0] array part
		if !ok {
			a := fmt.Sprintf("%s[0]", refAddr)
			refData, ok = resData[a]

			if ok {
				log.Debugf("reference does not specify a count: using resource %s for for %s.%s", a, d.Address, attr)
			}
		}

		if ok {
			found = true

			d.AddReference(attr, refData)
		}
	}

	return found
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
