package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var DefaultProviderRegion = "us-east-1"

func GetDefaultRefIDFunc(d *schema.ResourceData) []string {

	defaultRefs := []string{d.Get("urn").String()}

	arnAttr := "urn"
	if d.Get(arnAttr).Exists() {
		defaultRefs = append(defaultRefs, d.Get(arnAttr).String())
	}

	return defaultRefs
}

func GetSpecialContext(d *schema.ResourceData) map[string]interface{} {

	specialContexts := make(map[string]interface{})

	if strings.HasPrefix(d.Get("region").String(), "cn-") {
		specialContexts["isAWSChina"] = true
	}

	return specialContexts
}

func GetResourceRegion(resourceType string, v gjson.Result) string {
	// If a region key exists in the values use that

	return v.Get("region").String()

}

func ParseTags(resourceType string, v gjson.Result) map[string]string {
	tags := make(map[string]string)

	for k, v := range v.Get("tags").Map() {
		if k == "__defaults" {
			continue
		}
		tags[k] = v.String()
	}
	log.Debugf("tags %s", tags)
	return tags
}
