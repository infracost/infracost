package usage

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

// ResourceUsage represents a resource block in the usage file
type ResourceUsage struct {
	Name  string
	Items []*schema.UsageItem
}

func (r *ResourceUsage) Map() map[string]interface{} {
	m := make(map[string]interface{}, len(r.Items))
	for _, item := range r.Items {
		m[item.Key] = mapUsageItem(item)
	}

	return m
}

// MergeResourceUsage merge ResourceItem from src to r without overriding r
func (r *ResourceUsage) MergeResourceUsage(src *ResourceUsage) {
	if src == nil {
		return
	}

	destItemMap := make(map[string]*schema.UsageItem, len(r.Items))
	for _, item := range r.Items {
		destItemMap[item.Key] = item
	}

	for _, srcItem := range src.Items {
		destItem, ok := destItemMap[srcItem.Key]
		if !ok {
			destItem = &schema.UsageItem{Key: srcItem.Key}
			r.Items = append(r.Items, destItem)
		}

		if srcItem.ValueType == schema.SubResourceUsage {
			if srcItem.DefaultValue != nil {
				srcDefaultValue := srcItem.DefaultValue.(*ResourceUsage)
				if destItem.DefaultValue == nil {
					destItem.DefaultValue = &ResourceUsage{
						Name: srcDefaultValue.Name,
					}
				}
				d := destItem.DefaultValue.(*ResourceUsage)
				d.MergeResourceUsage(srcDefaultValue)
			}

			if srcItem.Value != nil {
				srcValue := srcItem.Value.(*ResourceUsage)
				if destItem.Value == nil {
					destItem.Value = destItem.DefaultValue
				}
				if destItem.Value == nil {
					destItem.Value = &ResourceUsage{
						Name: srcValue.Name,
					}
				}
				d := destItem.Value.(*ResourceUsage)
				d.MergeResourceUsage(srcValue)
			}
		} else if destItem.Value == nil {
			destItem.ValueType = srcItem.ValueType

			if srcItem.Description != "" {
				destItem.Description = srcItem.Description
			}

			if srcItem.DefaultValue != nil {
				destItem.DefaultValue = srcItem.DefaultValue
			}

			if srcItem.Value != nil {
				destItem.Value = srcItem.Value
			}
		}
	}
}

func mapUsageItem(item *schema.UsageItem) interface{} {
	if item.ValueType == schema.SubResourceUsage {
		m := make(map[string]interface{})

		if item.Value != nil {
			subResourceUsage := item.Value.(*ResourceUsage)
			for _, item := range subResourceUsage.Items {
				m[item.Key] = mapUsageItem(item)
			}
		}

		return m
	}

	return item.Value
}

func ResourceUsagesFromYAML(raw yamlv3.Node) ([]*ResourceUsage, error) {
	if len(raw.Content)%2 != 0 {
		// This error shouldn't really happen, the YAML lib flattens map node key and values into a single array
		// so this means the YAML map node is invalid but to be safe we add a log here in case it does.
		logging.Logger.Error().Msgf("YAML resource usage contents are not divisible by 2. Expected map node to have equal number of key and value nodes.")
		return []*ResourceUsage{}, errors.New("unexpected YAML format")
	}

	resourceUsages := make([]*ResourceUsage, 0, len(raw.Content)/2)

	for i := 0; i < len(raw.Content); i += 2 {
		resourceKeyNode := raw.Content[i]
		resourceValNode := raw.Content[i+1]

		if len(resourceValNode.Content)%2 != 0 {
			// This error shouldn't really happen, the YAML lib flattens map node key and values into a single array
			// so this means the YAML map node is invalid but to be safe we add a log here in case it does.
			logging.Logger.Error().Msgf("YAML resource value contents are not divisible by 2. Expected map node to have equal number of key and value nodes.")
			return resourceUsages, errors.New("unexpected YAML format")
		}

		resourceUsage := &ResourceUsage{
			Name:  resourceKeyNode.Value,
			Items: make([]*schema.UsageItem, 0, len(resourceValNode.Content)/2),
		}

		for i := 0; i < len(resourceValNode.Content); i += 2 {
			attrKeyNode := resourceValNode.Content[i]
			attrValNode := resourceValNode.Content[i+1]

			usageItem, err := usageItemFromYAML(attrKeyNode, attrValNode)
			if err != nil {
				return resourceUsages, err
			}

			resourceUsage.Items = append(resourceUsage.Items, usageItem)
		}

		resourceUsages = append(resourceUsages, resourceUsage)
	}

	return resourceUsages, nil
}

func resourceUsagesMap(resourceUsages []*ResourceUsage) map[string]*ResourceUsage {
	m := make(map[string]*ResourceUsage)

	for _, resourceUsage := range resourceUsages {
		m[resourceUsage.Name] = resourceUsage
	}

	return m
}

func ResourceUsagesToYAML(resourceUsages []*ResourceUsage) (yamlv3.Node, bool) {
	rootNode := yamlv3.Node{
		Kind: yamlv3.MappingNode,
	}

	rootNodeIsCommented := true

	for _, resourceUsage := range resourceUsages {
		if len(resourceUsage.Items) == 0 {
			continue
		}

		resourceKeyNode := &yamlv3.Node{
			Kind:  yamlv3.ScalarNode,
			Tag:   "!!str",
			Value: resourceUsage.Name,
		}

		resourceValNode := &yamlv3.Node{
			Kind: yamlv3.MappingNode,
		}

		resourceNodeIsCommented := true

		for _, item := range resourceUsage.Items {
			kind := yamlv3.ScalarNode
			content := make([]*yamlv3.Node, 0)

			itemNodeIsCommented := true
			rawValue := item.DefaultValue

			if item.Value != nil {
				rawValue = item.Value
				itemNodeIsCommented = false
				resourceNodeIsCommented = false
				rootNodeIsCommented = false
			}

			if item.ValueType == schema.SubResourceUsage {
				// If the value is a subresource, we need to add in any missing default sub-items
				// so they get rendered as comments
				subResourceUsage := &ResourceUsage{
					Name: item.Key,
				}
				if item.Value != nil {
					subResourceUsage = item.Value.(*ResourceUsage)
				}

				subResourceUsageMap := subResourceUsage.Map()

				if item.DefaultValue != nil {
					for _, defaultItem := range item.DefaultValue.(*ResourceUsage).Items {
						if _, ok := subResourceUsageMap[defaultItem.Key]; !ok {
							subResourceUsage.Items = append(subResourceUsage.Items, defaultItem)
						}
					}
				}

				subResourceValNode, allSubResourcesCommented := ResourceUsagesToYAML([]*ResourceUsage{subResourceUsage})

				if !allSubResourcesCommented {
					resourceNodeIsCommented = false
					rootNodeIsCommented = false
				}

				resourceValNode.Content = append(resourceValNode.Content, subResourceValNode.Content...)

				continue
			}

			var tag string
			var value string

			switch item.ValueType {
			case schema.Float64:
				tag = "!!float"

				// Float values might be represented as integers, so we need to make sure it's a float first
				var rawFloat float64
				switch f := rawValue.(type) {
				case float64:
					rawFloat = f
				case int64:
					rawFloat = float64(f)
				case int:
					rawFloat = float64(f)
				}

				// Format the float with as few decimal places as necessary
				value = strconv.FormatFloat(rawFloat, 'f', -1, 64)

				// If the float is a whole number then add at least one decimal place
				// so the YAML marshaller doesn't need to add an explicit !!float tag
				if value == fmt.Sprintf("%.f", rawFloat) {
					value = fmt.Sprintf("%s.0", value)
				}
			case schema.Int64:
				tag = "!!int"
				value = fmt.Sprintf("%d", rawValue)
			case schema.String:
				tag = "!!str"
				value = fmt.Sprintf("%s", rawValue)
			case schema.StringArray:
				tag = "!!seq"
				kind = yamlv3.SequenceNode
				for _, item := range rawValue.([]string) {
					content = append(content, &yamlv3.Node{
						Kind:  yamlv3.ScalarNode,
						Tag:   "!!str",
						Value: item,
					})
				}
			case schema.Bool:
				tag = "!!bool"
				value = fmt.Sprintf("%v", rawValue)
			}

			itemKeyNode := &yamlv3.Node{
				Kind:  yamlv3.ScalarNode,
				Tag:   "!!str",
				Value: item.Key,
			}
			if itemNodeIsCommented {
				markNodeAsComment(itemKeyNode)
			}

			itemValNode := &yamlv3.Node{
				Kind:        kind,
				Tag:         tag,
				Value:       value,
				Content:     content,
				LineComment: item.Description,
			}

			resourceValNode.Content = append(resourceValNode.Content, itemKeyNode)
			resourceValNode.Content = append(resourceValNode.Content, itemValNode)
		}

		if resourceNodeIsCommented {
			markNodeAsComment(resourceKeyNode)
		}

		rootNode.Content = append(rootNode.Content, resourceKeyNode)
		rootNode.Content = append(rootNode.Content, resourceValNode)
	}

	return rootNode, rootNodeIsCommented
}

// usageItemFromYAML item turns a YAML key node and a YAML value node into a *schema.UsageItem. This function supports recursion
// to allow for YAML map nodes to be parsed into nested sets of schema.UsageItem
//
// e.g. given:
//
//	keyNode: &yaml.Node{
//		Value: "testKey",
//	}
//
//	valNode: &yaml.Node{
//		Kind: yaml.MappingNode,
//		Content: []*yaml.Node{
//			&yaml.Node{Value: "prop1"},
//			&yaml.Node{Value: "test"},
//			&yaml.Node{Value: "prop2"},
//			&yaml.Node{Value: "test2"},
//			&yaml.Node{Value: "prop3"},
//			&yaml.Node{
//				Kind: yaml.MappingNode,
//				Content: []*yaml.Node{
//					&yaml.Node{Value: "nested1"},
//					&yaml.Node{Value: "test3"},
//				},
//			},
//		},
//	}
//
// usageItemFromYAML will return:
//
//	UsageItem{
//			Key:          "testKey",
//			Value: []*UsageItem{
//				{
//					Key: "prop1",
//					Value: "test",
//				},
//				{
//					Key: "prop2",
//					Value: "test2",
//				},
//				{
//					Key: "prop3",
//					Value: []*UsageItem{
//						{
//							Key: "nested1",
//							Value: "test3",
//						},
//					},
//				},
//			},
//		}
func usageItemFromYAML(keyNode *yamlv3.Node, valNode *yamlv3.Node) (*schema.UsageItem, error) {
	if keyNode == nil || valNode == nil {
		logging.Logger.Error().Msgf("YAML contains nil key or value node")
		return nil, errors.New("unexpected YAML format")
	}

	var value interface{}
	var usageValueType schema.UsageVariableType

	if valNode.ShortTag() == "!!map" {
		usageValueType = schema.SubResourceUsage

		if len(valNode.Content)%2 != 0 {
			// This error shouldn't really happen, the YAML lib flattens map node key and values into a single array
			// so this means the YAML map node is invalid but to be safe we add a log here in case it does.
			logging.Logger.Error().Msgf("YAML value map node contents are not divisible by 2. Expected map node to have equal number of key and value nodes.")
			return nil, errors.New("unexpected YAML format")
		}

		items := make([]*schema.UsageItem, 0, len(valNode.Content)/2)

		for i := 0; i < len(valNode.Content); i += 2 {
			mapKeyNode := valNode.Content[i]
			mapValNode := valNode.Content[i+1]

			mapUsageItem, err := usageItemFromYAML(mapKeyNode, mapValNode)
			if err != nil {
				return nil, err
			}

			items = append(items, mapUsageItem)
		}

		value = &ResourceUsage{
			Name:  keyNode.Value,
			Items: items,
		}
	} else {
		err := valNode.Decode(&value)
		if err != nil {
			logging.Logger.Error().Msgf("Unable to decode YAML value")
			return nil, errors.New("unexpected YAML format")
		}

		switch valNode.ShortTag() {
		case "!!int":
			usageValueType = schema.Int64
		case "!!float":
			usageValueType = schema.Float64
		case "!!bool":
			usageValueType = schema.Bool
		default:
			usageValueType = schema.String
		}
	}

	return &schema.UsageItem{
		Key:         keyNode.Value,
		ValueType:   usageValueType,
		Value:       value,
		Description: valNode.LineComment,
	}, nil
}
