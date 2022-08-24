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
	"ibm_tg_gateway":                {"transit.gateway", []string{}, nil},
	"ibm_is_floating_ip":            {"is.floating-ip", []string{}, nil},
	"ibm_is_flow_log":               {"is.flow-log-collector", []string{}, nil},
	"ibm_cloudant":                  {"cloudantnosqldb", []string{}, nil},
	"ibm_pi_instance":               {"power-iaas", []string{}, nil},
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
