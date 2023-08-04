package aws

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/providers/terraform/provider_schemas"
	"github.com/infracost/infracost/internal/schema"
)

type parseTagFunc func(baseTags map[string]string, r *schema.ResourceData)

var tagProviders = map[string]parseTagFunc{
	"aws_instance":          parseInstanceTags,
	"aws_autoscaling_group": parseAutoScalingTags,
}

func ParseTags(defaultTags *map[string]string, r *schema.ResourceData) *map[string]string {
	_, supportsTags := provider_schemas.AWSTagsSupport[r.Type]
	_, supportsTagBlock := provider_schemas.AWSTagBlockSupport[r.Type]

	rTags := r.Get("tags").Map()
	if !supportsTags && !supportsTagBlock && len(rTags) == 0 {
		return nil
	}

	tags := make(map[string]string)

	_, supportsDefaultTags := provider_schemas.AWSTagsAllSupport[r.Type]
	if supportsDefaultTags && defaultTags != nil {
		for k, v := range *defaultTags {
			tags[k] = v
		}
	}

	if supportsTagBlock {
		for _, el := range r.Get("tag").Array() {
			k := el.Get("key").String()
			if k == "" {
				continue
			}

			propagate := el.Get("propagate_at_launch")
			if propagate.Exists() && !propagate.Bool() {
				continue
			}

			tags[k] = el.Get("value").String()
		}
	}

	for k, v := range rTags {
		tags[k] = v.String()
	}

	if f, ok := tagProviders[r.Type]; ok {
		f(tags, r)
	}

	return &tags
}

func parseAutoScalingTags(tags map[string]string, r *schema.ResourceData) {
	tagSpecifications(r, func(resourceType string, specs map[string]gjson.Result) {
		if resourceType == "instance" {
			for k, v := range specs {
				tags[k] = v.String()
			}
		}
	})
}

func parseInstanceTags(tags map[string]string, r *schema.ResourceData) {
	for k, v := range r.Get("volume_tags").Map() {
		tags[fmt.Sprintf("volume_tags.%s", k)] = v.String()
	}

	tagSpecifications(r, func(resourceType string, specs map[string]gjson.Result) {
		if resourceType == "instance" {
			for k, v := range specs {
				tags[k] = v.String()
			}
		}

		if resourceType == "volume" {
			for k, v := range specs {
				tags[fmt.Sprintf("volume_tags.%s", k)] = v.String()
			}
		}
	})
}

func tagSpecifications(r *schema.ResourceData, f func(resourceType string, specs map[string]gjson.Result)) {
	for key, data := range r.ReferencesMap {
		if strings.HasPrefix(key, "launch_template") {
			launchTemplate := data[0]
			for _, s := range launchTemplate.Get("tag_specifications").Array() {
				resourceType := s.Get("resource_type")
				f(resourceType.String(), s.Get("tags").Map())
			}

			break
		}
	}
}
