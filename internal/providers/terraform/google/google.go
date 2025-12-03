package google

import (
	"maps"

	"github.com/infracost/infracost/internal/providers/terraform/provider_schemas"
	"github.com/infracost/infracost/internal/schema"
)

var DefaultProviderRegion = "us-central1"

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

func GetSpecialContext(d *schema.ResourceData) map[string]any {
	return map[string]any{}
}

func GetResourceRegion(d *schema.ResourceData) string {
	v := d.RawValues

	if v.Get("region").Exists() && v.Get("region").String() != "" {
		return v.Get("region").String()
	}

	return ""
}

func ParseTags(r *schema.ResourceData, externalTags, defaultLabels map[string]string) (map[string]string, []string) {

	_, supportsLabels := provider_schemas.GoogleLabelsSupport[r.Type]
	rLabels := r.Get("labels").Map()

	_, supportsUserLabels := provider_schemas.GoogleUserLabelsSupport[r.Type]
	rUserLabels := r.Get("user_labels").Map()

	_, supportsSettingsUserLabels := provider_schemas.GoogleSettingsUserLabelsSupport[r.Type]
	rSettingsUserLabels := r.Get("settings.0.user_labels").Map()

	missingForLabels := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "labels")
	missingForUserLabels := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "user_labels")
	missingForSettingsUserLabels := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "settings.0.user_labels")
	missing := append(append(missingForLabels, missingForUserLabels...), missingForSettingsUserLabels...)

	if !supportsLabels && len(rLabels) == 0 &&
		!supportsUserLabels && len(rUserLabels) == 0 &&
		!supportsSettingsUserLabels && len(rSettingsUserLabels) == 0 {
		return nil, missing
	}

	tags := make(map[string]string)

	maps.Copy(tags, defaultLabels)
	for k, v := range rLabels {
		tags[k] = v.String()
	}
	for k, v := range rUserLabels {
		tags[k] = v.String()
	}
	for k, v := range rSettingsUserLabels {
		tags[k] = v.String()
	}
	maps.Copy(tags, externalTags)

	return tags, missing
}
