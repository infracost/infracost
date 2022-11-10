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
	"ibm_tg_gateway":                {"f38a4da0-c353-11e9-83b6-a36a57a97a06", []string{}, nil},
	"ibm_is_floating_ip":            {"is.floating-ip", []string{}, nil},
	"ibm_is_flow_log":               {"is.flow-log-collector", []string{}, nil},
	"ibm_cloudant":                  {"cloudant", []string{}, nil},
	"ibm_pi_instance":               {"abd259f0-9990-11e8-acc8-b9f54a8f1661", []string{}, nil},
	"ibm_cos_bucket":                {"Cloud Object Storage Bucket", []string{}, nil},
	"ibm_is_lb":                     {"is.load-balancer", []string{}, nil},
	"ibm_is_public_gateway":         {"is.public-gateway", []string{"ibm_is_floating_ip"}, nil},
	"kms":                           {"ee41347f-b18e-4ca6-bf80-b5467c63f9a6", []string{}, nil},
	"cloud-object-storage":          {"dff97f5c-bc5e-4455-b470-411c3edbe49c", []string{}, nil},
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
