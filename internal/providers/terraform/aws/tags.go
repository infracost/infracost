package aws

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/providers/terraform/provider_schemas"
	"github.com/infracost/infracost/internal/schema"
)

type parseTagFunc func(baseTags, defaultTags map[string]string, r *schema.ResourceData, config TagParsingConfig)

var tagProviders = map[string]parseTagFunc{
	"aws_instance":          parseInstanceTags,
	"aws_autoscaling_group": parseAutoScalingTags,
	"aws_launch_template":   parseLaunchTemplateTags,
}

func parseLaunchTemplateTags(tags, defaultTags map[string]string, r *schema.ResourceData, config TagParsingConfig) {
	for _, s := range r.Get("tag_specifications").Array() {
		for k, v := range s.Get("tags").Map() {
			tags[fmt.Sprintf("tag_specifications.%s", k)] = v.String()
		}
	}
}

// TagParsingConfig defines options that can be used to configure the ParseTags function.
type TagParsingConfig struct {
	// PropagateDefaultsToVolumeTags specifies whether default provider tags should be
	// propagated to volume tags. This feature was added in Terraform AWS provider
	// version 5.39.0.
	PropagateDefaultsToVolumeTags bool
}

type TagPropagationConfig struct {
	Attribute string
	To        string
	RefMap    map[string]string
	Requires  []string
}

// ExpectedPropagations describe known tag propagation configurations
var ExpectedPropagations = map[string]TagPropagationConfig{
	"aws_ecs_service": {
		Attribute: "propagate_tags",
		To:        "task",
		RefMap: map[string]string{
			"TASK_DEFINITION": "task_definition",
			"SERVICE":         "", // empty string means self-reference
		},
	},
	"aws_scheduler_schedule": {
		Attribute: "ecs_parameters.0.propagate_tags",
		To:        "task",
		RefMap: map[string]string{
			"TASK_DEFINITION": "ecs_parameters.0.task_definition_arn",
		},
	},
	"aws_batch_job_definition": {
		Attribute: "propagate_tags",
		To:        "task",
		RefMap: map[string]string{
			"true": "", // empty string means self-reference
		},
	},
	"aws_dynamodb_table": {
		Attribute: "replica.0.propagate_tags",
		To:        "replica",
		RefMap: map[string]string{
			"true": "", // empty string means self-reference
		},
		Requires: []string{"replica.0.region_name"},
	},
	"aws_pipes_pipe": {
		Attribute: "target_parameters.0.ecs_task_parameters.0.propagate_tags",
		To:        "task",
		RefMap: map[string]string{
			"TASK_DEFINITION": "target_parameters.0.ecs_task_parameters.0.task_definition_arn",
		},
	},
	"aws_cloudwatch_event_target": {
		Attribute: "ecs_target.0.propagate_tags",
		To:        "task",
		RefMap: map[string]string{
			"TASK_DEFINITION": "ecs_target.0.task_definition_arn",
		},
		Requires: []string{"ecs_target.0.task_definition_arn"},
	},
}

func ParseTags(externalTags, defaultTags map[string]string, r *schema.ResourceData, config TagParsingConfig) (map[string]string, []string) {
	_, supportsTags := provider_schemas.AWSTagsSupport[r.Type]
	_, supportsTagBlock := provider_schemas.AWSTagBlockSupport[r.Type]

	var missing []string
	tagsMissing := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "tags")
	missing = append(missing, tagsMissing...)
	tagsAllMissing := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "tags_all")
	missing = append(missing, tagsAllMissing...)

	rTagBlock := r.Get("tags")
	var rTags map[string]gjson.Result

	// If the tags attribute is a list, we need to handle it differently - this is the case with AutoScalingGroups
	// where the tags are a slice of objects with details about propagation. When it is a list, we should parse it as such
	switch t := rTagBlock.Value().(type) {
	case []any:
		rTags = make(map[string]gjson.Result, len(t))
		for _, el := range t {
			tag, ok := el.(map[string]any)
			if !ok {
				// This should never happen, but if it does, we should skip it
				continue
			}
			k, ok := tag["key"].(string)
			if !ok {
				// the key was forgotten, so we should skip it
				continue
			}
			v, ok := tag["value"].(string)
			if !ok {
				// the value was forgotten, so we should skip it
				continue
			}
			rTags[k] = gjson.Parse(fmt.Sprintf(`"%s"`, v))
		}
	default:
		rTags = rTagBlock.Map()
	}

	rTagsAll := r.Get("tags_all").Map()
	if !supportsTags && !supportsTagBlock && len(rTags) == 0 && len(rTagsAll) == 0 {
		return nil, missing
	}

	tags := make(map[string]string)

	_, supportsDefaultTags := provider_schemas.AWSTagsAllSupport[r.Type]
	if supportsDefaultTags && defaultTags != nil {
		keysAndValues := make([]string, 0, len(defaultTags)*2)
		for k, v := range defaultTags {
			tags[k] = v
			keysAndValues = append(keysAndValues, k, v)
		}

		sort.Strings(keysAndValues)

		h := sha256.New()
		for _, s := range keysAndValues {
			h.Write([]byte(s))
		}

		checksum := hex.EncodeToString(h.Sum(nil))
		r.Metadata["defaultTagsChecksum"] = gjson.Parse(fmt.Sprintf(`"%s"`, checksum))
	}

	if supportsTagBlock {
		tagBlockMissing := schema.ExtractMissingVarsCausingMissingAttributeKeys(r, "tag")
		missing = append(missing, tagBlockMissing...)
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

	// tags_all is only set on plan.json runs, so we can prefer them over our own calculations
	for k, v := range rTagsAll {
		tags[k] = v.String()
	}

	// external tags (e.g. yor)
	maps.Copy(tags, externalTags)

	if f, ok := tagProviders[r.Type]; ok {
		f(tags, defaultTags, r, config)
	}

	return tags, missing
}

func parseAutoScalingTags(tags, defaultTags map[string]string, r *schema.ResourceData, config TagParsingConfig) {
	referencedTagSpecifications(r, func(resourceType string, specs map[string]gjson.Result) {
		if resourceType == "instance" {
			for k, v := range specs {
				tags[k] = v.String()
			}
		}
	})
}

func parseInstanceTags(tags, defaultTags map[string]string, r *schema.ResourceData, config TagParsingConfig) {
	if config.PropagateDefaultsToVolumeTags && len(defaultTags) > 0 {
		// when propagating default tags, we add them to volume_tags if they already exist
		// otherwise they are propogated to directly to the block devices.

		if r.Get("volume_tags").Exists() {
			for k, v := range defaultTags {
				tags[fmt.Sprintf("volume_tags.%s", k)] = v
			}
		} else {
			// a root_block_device is assumed even if not explicitly defined
			for k, v := range defaultTags {
				tags[fmt.Sprintf("root_block_device.%s", k)] = v
			}
			for i := range r.Get("ebs_block_device").Array() {
				for k, v := range defaultTags {
					tags[fmt.Sprintf("ebs_block_device[%d].%s", i, k)] = v
				}
			}
		}
	}

	if rbd := r.Get("root_block_device"); rbd.Exists() {
		for k, v := range rbd.Get("0.tags").Map() {
			tags[fmt.Sprintf("root_block_device.%s", k)] = v.String()
		}
	}

	for i, vol := range r.Get("ebs_block_device").Array() {
		for k, v := range vol.Get("tags").Map() {
			tags[fmt.Sprintf("ebs_block_device[%d].%s", i, k)] = v.String()
		}
	}

	for k, v := range r.Get("volume_tags").Map() {
		tags[fmt.Sprintf("volume_tags.%s", k)] = v.String()
	}

	referencedTagSpecifications(r, func(resourceType string, specs map[string]gjson.Result) {
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

func referencedTagSpecifications(r *schema.ResourceData, f func(resourceType string, specs map[string]gjson.Result)) {
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
