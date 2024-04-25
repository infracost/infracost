package ibm

import (
	"encoding/json"
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

var DefaultProviderRegion = "us-south"

func GetDefaultRefIDFunc(d *schema.ResourceData) []string {

	defaultRefs := []string{d.Get("id").String()}

	if d.Get("self_link").Exists() {
		defaultRefs = append(defaultRefs, d.Get("self_link").String())
	}

	return defaultRefs
}

func DefaultCloudResourceIDFunc(d *schema.ResourceData) []string {
	return []string{}
}

func GetSpecialContext(d *schema.ResourceData) map[string]interface{} {
	return map[string]interface{}{}
}

func GetResourceRegion(resourceType string, v gjson.Result) string {
	return ""
}

func ParseTags(resourceType string, v gjson.Result) map[string]string {
	tags := make(map[string]string)
	for k, v := range v.Get("labels").Map() {
		tags[k] = v.String()
	}
	return tags
}

type catalogMetadata struct {
	serviceId      string
	childResources []string
	configuration  map[string]any
}

// Map between terraform type and global catalog id. For ibm_resource_instance, the service
// field already matches the global catalog id, so they do not need to be mapped. eg: "kms"
var globalCatalogServiceId = map[string]catalogMetadata{
	"ibm_is_vpc":                    {"is.vpc", []string{"ibm_is_flow_log"}, nil},
	"ibm_container_vpc_cluster":     {"containers-kubernetes", []string{}, nil},
	"ibm_container_vpc_worker_pool": {"containers-kubernetes", []string{}, nil},
	"ibm_is_instance":               {"is.instance", []string{"ibm_is_ssh_key", "ibm_is_floating_ip"}, nil},
	"ibm_is_volume":                 {"is.volume", []string{}, nil},
	"ibm_is_vpn_gateway":            {"is.vpn", []string{}, nil},
	"ibm_tg_gateway":                {"f38a4da0-c353-11e9-83b6-a36a57a97a06", []string{}, nil},
	"ibm_is_floating_ip":            {"is.floating-ip", []string{}, nil},
	"ibm_is_flow_log":               {"is.flow-log-collector", []string{}, nil},
	"ibm_cloudant":                  {"Cloudant", []string{}, nil},
	"ibm_pi_instance":               {"Power Systems Virtual Server", []string{}, nil},
	"ibm_pi_volume":                 {"Power Systems Storage Volume", []string{}, nil},
	"ibm_cos_bucket":                {"Object Storage Bucket", []string{}, nil},
	"ibm_is_lb":                     {"is.load-balancer", []string{}, nil},
	"ibm_is_public_gateway":         {"is.public-gateway", []string{"ibm_is_floating_ip"}, nil},
	"kms":                           {"ee41347f-b18e-4ca6-bf80-b5467c63f9a6", []string{}, nil},
	"cloud-object-storage":          {"dff97f5c-bc5e-4455-b470-411c3edbe49c", []string{}, nil},
	"roks":                          {"containers.kubernetes.cluster.roks", []string{}, nil},
	"pm-20":                         {"51c53b72-918f-4869-b834-2d99eb28422a", []string{}, nil},
	"data-science-experience":       {"39ba9d4c-b1c5-4cc3-a163-38b580121e01", []string{}, nil},
	"discovery":                     {"76b7bf22-b443-47db-b3db-066ba2988f47", []string{}, nil},
	"wx":                            {"51c53b72-918f-4869-b834-2d99eb28422a", []string{}, nil},
	"conversation":                  {"7045626d-55e3-4418-be11-683a26dbc1e5", []string{}, nil},
	"aiopenscale":                   {"2ad019f3-0fd6-4c25-966d-f3952481a870", []string{}, nil},
	"sysdig-monitor":                {"090c2c10-8c38-11e8-bec2-493df9c49eb8", []string{}, nil},
	"sysdig-secure":                 {"e831e900-82d6-11ec-95c5-c12c5a5d9687", []string{}, nil},
	"logdna":                        {"e13e1860-959c-11e8-871e-ad157af61ad7", []string{}, nil},
}

func SetCatalogMetadata(d *schema.ResourceData, resourceType string, config map[string]any) {
	metadata := make(map[string]gjson.Result)
	var properties gjson.Result
	var serviceId string = resourceType
	var childResources []string

	catalogEntry, isPresent := globalCatalogServiceId[resourceType]
	if isPresent {
		serviceId = catalogEntry.serviceId
		childResources = catalogEntry.childResources
	}

	configString, err := json.Marshal(config)
	if err != nil {
		configString = []byte("{}")
	}

	if len(childResources) > 0 {
		childResourcesString, err := json.Marshal(childResources)
		if err != nil {
			childResourcesString = []byte("[]")
		}

		properties = gjson.Result{
			Type: gjson.JSON,
			Raw:  fmt.Sprintf(`{"serviceId": "%s" , "childResources": %s, "configuration": %s}`, serviceId, childResourcesString, configString),
		}
	} else {
		properties = gjson.Result{
			Type: gjson.JSON,
			Raw:  fmt.Sprintf(`{"serviceId": "%s", "configuration": %s}`, serviceId, configString),
		}
	}

	metadata["catalog"] = properties
	d.Metadata = metadata
}
