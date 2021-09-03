package crossplane

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13.
var infracostProviderNames = []string{"infracost", "registry.terraform.io/infracost/infracost"}
var defaultProviderRegions = map[string]string{
	"aws":     "us-east-1",
	"google":  "us-central1",
	"azurerm": "eastus",
}

// ARN attribute mapping for resources that don't have a standard 'arn' attribute
var arnAttributeMap = map[string]string{
	"aws_cloudwatch_dashboard":     "dashboard_arn",
	"aws_db_snapshot":              "db_snapshot_arn",
	"aws_db_cluster_snapshot":      "db_cluster_snapshot_arn",
	"aws_ecs_service":              "id",
	"aws_neptune_cluster_snapshot": "db_cluster_snapshot_arn",
	"aws_docdb_cluster_snapshot":   "db_cluster_snapshot_arn",
	"aws_dms_certificate":          "certificate_arn",
	"aws_dms_endpoint":             "endpoint_arn",
	"aws_dms_replication_instance": "replication_instance_arn",
	"aws_dms_replication_task":     "replication_task_arn",
}

// Parser ...
type Parser struct {
	ctx *config.ProjectContext
}

// NewParser ...
func NewParser(ctx *config.ProjectContext) *Parser {
	return &Parser{ctx}
}

func (p *Parser) createResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	registryMap := GetResourceRegistryMap()

	if registryItem, ok := (*registryMap)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				Tags:         d.Tags,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
			}
		}
		res := registryItem.RFunc(d, u)
		if res != nil {
			res.ResourceType = d.Type
			res.Tags = d.Tags
			return res
		}
	}
	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		IsSkipped:    true,
		SkipMessage:  "This resource is not currently supported",
	}
}

func (p *Parser) parseJSON(j []byte, usage map[string]*schema.UsageData) ([]*schema.Resource, []*schema.Resource, error) {
	baseResources := p.loadUsageFileResources(usage)

	if !gjson.ValidBytes(j) {
		return baseResources, baseResources, errors.New("invalid JSON")
	}

	parsed := gjson.ParseBytes(j)

	log.Infof("parsed: %s", parsed)

	// TODO: Revisit. Next action Item to work.

	return nil, nil, nil

	providerConf := parsed.Get("configuration.provider_config")
	conf := parsed.Get("configuration.root_module")
	vars := parsed.Get("variables")

	// TODO: Do we support pastResources with CrossPlane?
	// pastResources := p.parseJSONResources(true, baseResources, usage, parsed, providerConf, conf, vars)
	var pastResources []*schema.Resource
	resources := p.parseJSONResources(false, baseResources, usage, parsed, providerConf, conf, vars)

	return pastResources, resources, nil
}

func (p *Parser) parseJSONResources(parsePrior bool, baseResources []*schema.Resource, usage map[string]*schema.UsageData, parsed, providerConf, conf, vars gjson.Result) []*schema.Resource {
	var resources []*schema.Resource
	resources = append(resources, baseResources...)
	var vals gjson.Result
	// TODO: Revisit
	if parsePrior {
		vals = parsed.Get("prior_state.values.root_module")
	} else {
		vals = parsed.Get("planned_values.root_module")
		if !vals.Exists() {
			vals = parsed.Get("values.root_module")
		}
	}

	resData := p.parseResourceData(providerConf, vals, conf, vars)

	// p.parseReferences(resData, conf)
	// p.loadInfracostProviderUsageData(usage, resData)
	// p.stripDataResources(resData)

	for _, d := range resData {
		var usageData *schema.UsageData

		if ud := usage[d.Address]; ud != nil {
			usageData = ud
		} else if strings.HasSuffix(d.Address, "]") {
			lastIndexOfOpenBracket := strings.LastIndex(d.Address, "[")

			if arrayUsageData := usage[fmt.Sprintf("%s[*]", d.Address[:lastIndexOfOpenBracket])]; arrayUsageData != nil {
				usageData = arrayUsageData
			}
		}
		if r := p.createResource(d, usageData); r != nil {
			resources = append(resources, r)
		}
	}

	return resources
}

func (p *Parser) parseResourceData(providerConf, planVals gjson.Result, conf gjson.Result, vars gjson.Result) map[string]*schema.ResourceData {
	resources := make(map[string]*schema.ResourceData)

	// for _, r := range planVals.Get("resources").Array() {
	// 	t := r.Get("type").String()
	// 	provider := r.Get("provider_name").String()
	// 	addr := r.Get("address").String()
	// 	v := r.Get("values")

	// 	resConf := getConfJSON(conf, addr)

	// 	// Try getting the region from the ARN
	// 	region := resourceRegion(t, v)

	// 	// Otherwise use region from the provider conf
	// 	if region == "" {
	// 		region = providerRegion(addr, providerConf, vars, t, resConf)
	// 	}

	// 	v = schema.AddRawValue(v, "region", region)

	// 	tags := parseTags(t, v)

	// 	resources[addr] = schema.NewResourceData(t, provider, addr, tags, v)
	// }

	// Recursively add any resources for child modules
	for _, m := range planVals.Get("child_modules").Array() {
		for addr, d := range p.parseResourceData(providerConf, m, conf, vars) {
			resources[addr] = d
		}
	}

	return resources
}

func (p *Parser) loadUsageFileResources(u map[string]*schema.UsageData) []*schema.Resource {
	resources := make([]*schema.Resource, 0)
	for k, v := range u {
		for _, t := range GetUsageOnlyResources() {
			if strings.HasPrefix(k, fmt.Sprintf("%s.", t)) {
				d := schema.NewResourceData(t, "global", k, map[string]string{}, gjson.Result{})
				if r := p.createResource(d, v); r != nil {
					resources = append(resources, r)
				}
			}
		}
	}
	return resources
}
