package azure

import (
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

func GetSpecialContext(d *schema.ResourceData) map[string]interface{} {
	return map[string]interface{}{}
}

func ParseTags(r *schema.ResourceData) *map[string]string {
	_, supportsTags := provider_schemas.AzureTagsSupport[r.Type]
	rTags := r.Get("tags").Map()
	if !supportsTags && len(rTags) == 0 {
		return nil
	}

	tags := make(map[string]string)
	for k, v := range rTags {
		tags[k] = v.String()
	}
	return &tags
}
