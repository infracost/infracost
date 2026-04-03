package azure

import (
	"maps"

	"github.com/infracost/infracost/internal/providers/terraform/provider_schemas"
	"github.com/infracost/infracost/internal/schema"
)

var DefaultProviderRegion = "eastus"

func GetDefaultRefIDFunc(d *schema.ResourceData) []string {
	return []string{d.Get("id").String()}
}

func DefaultCloudResourceIDFunc(d *schema.ResourceData) []string {
	return []string{}
}

func GetSpecialContext(d *schema.ResourceData) map[string]any {
	return map[string]any{}
}

func ParseTags(externalTags map[string]string, r *schema.ResourceData) (map[string]string, []string) {
	_, supportsTags := provider_schemas.AzureTagsSupport[r.Type]
	rTags := r.Get("tags").Map()
	missing := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "tags")
	if !supportsTags && len(rTags) == 0 {
		return nil, missing
	}
	tags := make(map[string]string)
	for k, v := range rTags {
		tags[k] = v.String()
	}
	maps.Copy(tags, externalTags)
	return tags, missing
}
