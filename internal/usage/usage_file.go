package usage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	yamlv3 "gopkg.in/yaml.v3"
)

const minUsageFileVersion = "0.1"
const maxUsageFileVersion = "0.1"
const commentMark = "00__"

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

func mapUsageItem(item *schema.UsageItem) interface{} {
	if item.ValueType == schema.Items {
		subItems := item.Value.([]*schema.UsageItem)
		m := make(map[string]interface{}, len(subItems))
		for _, item := range subItems {
			m[item.Key] = mapUsageItem(item)
		}

		return m
	}

	return item.Value
}

func resourceUsagesMap(resourceUsages []*ResourceUsage) map[string]*ResourceUsage {
	m := make(map[string]*ResourceUsage)

	for _, resourceUsage := range resourceUsages {
		m[resourceUsage.Name] = resourceUsage
	}

	return m
}

type UsageFile struct { // nolint:revive
	Version          string           `yaml:"version"`
	RawResourceUsage yamlv3.Node      `yaml:"resource_usage"`
	ResourceUsages   []*ResourceUsage `yaml:"-"`
}

func LoadUsageFile(path string, createIfNotExist bool) (*UsageFile, error) {
	if createIfNotExist {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Debug("Specified usage file does not exist. It will be created")

			return CreateBlankUsageFile(path)
		}
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return &UsageFile{}, errors.Wrapf(err, "Error reading usage file")
	}

	return LoadUsageFileFromString(string(contents))
}

func CreateBlankUsageFile(path string) (*UsageFile, error) {
	usageFile := &UsageFile{
		Version: maxUsageFileVersion,
		RawResourceUsage: yamlv3.Node{
			Kind: yamlv3.MappingNode,
		},
	}
	d, err := yamlv3.Marshal(usageFile)
	if err != nil {
		return usageFile, errors.Wrapf(err, "Error creating usage file")
	}
	err = ioutil.WriteFile(path, d, 0600)
	if err != nil {
		return usageFile, errors.Wrapf(err, "Error creating usage file")
	}

	return usageFile, nil
}

func LoadUsageFileFromString(s string) (*UsageFile, error) {
	usageFile := &UsageFile{}

	err := yamlv3.Unmarshal([]byte(s), usageFile)
	if err != nil {
		return usageFile, errors.Wrap(err, "Error parsing usage YAML")
	}

	if !usageFile.checkVersion() {
		return usageFile, fmt.Errorf("Invalid usage file version. Supported versions are %s ≤ x ≤ %s", minUsageFileVersion, maxUsageFileVersion)
	}

	err = usageFile.loadResourceUsages()
	if err != nil {
		return usageFile, errors.Wrap(err, "Error loading YAML file")
	}

	return usageFile, nil
}

func (u *UsageFile) WriteToPath(path string) error {
	var resourceUsageNodeIsCommented bool
	u.RawResourceUsage, resourceUsageNodeIsCommented = resourceUsagesToYAML(u.ResourceUsages)

	var buf bytes.Buffer
	yamlEncoder := yamlv3.NewEncoder(&buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(u)
	if err != nil {
		return err
	}

	b := buf.Bytes()
	b = replaceCommentMarks(b, resourceUsageNodeIsCommented)
	err = ioutil.WriteFile(path, b, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (u *UsageFile) ToUsageDataMap() map[string]*schema.UsageData {
	m := make(map[string]*schema.UsageData)

	for _, resourceUsage := range u.ResourceUsages {
		m[resourceUsage.Name] = schema.NewUsageData(resourceUsage.Name, schema.ParseAttributes(resourceUsage.Map()))
	}

	return m
}

func (u *UsageFile) checkVersion() bool {
	v := u.Version
	if !strings.HasPrefix(u.Version, "v") {
		v = "v" + u.Version
	}
	return semver.Compare(v, "v"+minUsageFileVersion) >= 0 && semver.Compare(v, "v"+maxUsageFileVersion) <= 0
}

func (u *UsageFile) loadResourceUsages() error {
	if len(u.RawResourceUsage.Content)%2 != 0 {
		log.Errorf("YAML resource usage contents are not divisible by 2")
		return errors.New("unexpected YAML format")
	}

	u.ResourceUsages = make([]*ResourceUsage, 0, len(u.RawResourceUsage.Content)/2)

	for i := 0; i < len(u.RawResourceUsage.Content); i += 2 {
		resourceKeyNode := u.RawResourceUsage.Content[i]
		resourceValNode := u.RawResourceUsage.Content[i+1]

		if len(resourceValNode.Content)%2 != 0 {
			log.Errorf("YAML resource value contents are not divisible by 2")
			return errors.New("unexpected YAML format")
		}

		resourceUsage := &ResourceUsage{
			Name:  resourceKeyNode.Value,
			Items: make([]*schema.UsageItem, 0, len(resourceValNode.Content)/2),
		}

		for i := 0; i < len(resourceValNode.Content); i += 2 {
			attrKeyNode := resourceValNode.Content[i]
			attrValNode := resourceValNode.Content[i+1]

			usageItem, err := yamlToUsageItem(attrKeyNode, attrValNode)
			if err != nil {
				return err
			}

			resourceUsage.Items = append(resourceUsage.Items, usageItem)
		}

		u.ResourceUsages = append(u.ResourceUsages, resourceUsage)
	}

	return nil
}

// yamlToUsageItem item turns a YAML key node and a YAML value node into a *schema.UsageItem. This function supports recursion
// to allow for YAML map nodes to be parsed into nested sets of schema.UsageItem
//
// e.g. given:
//
//		keyNode: &yaml.Node{
//			Value: "testKey",
//		}
//
//		valNode: &yaml.Node{
//			Kind: yaml.MappingNode,
//			Content: []*yaml.Node{
//				&yaml.Node{Value: "prop1"},
//				&yaml.Node{Value: "test"},
//				&yaml.Node{Value: "prop2"},
//				&yaml.Node{Value: "test2"},
//				&yaml.Node{Value: "prop3"},
//				&yaml.Node{
//					Kind: yaml.MappingNode,
//					Content: []*yaml.Node{
//						&yaml.Node{Value: "nested1"},
//						&yaml.Node{Value: "test3"},
//					},
//				},
//			},
//		}
//
// yamlToUsageItem will return:
//
// 		UsageItem{
//				Key:          "testKey",
//				Value: []*UsageItem{
//					{
//						Key: "prop1",
//						Value: "test",
//					},
//					{
//						Key: "prop2",
//						Value: "test2",
//					},
//					{
//						Key: "prop3",
//						Value: []*UsageItem{
//							{
//								Key: "nested1",
//								Value: "test3",
//							},
//						},
//					},
//				},
//			}
//
func yamlToUsageItem(keyNode *yamlv3.Node, valNode *yamlv3.Node) (*schema.UsageItem, error) {
	if keyNode == nil || valNode == nil {
		log.Errorf("YAML contains nil key or value node")
		return nil, errors.New("unexpected YAML format")
	}

	var value interface{}
	var usageValueType schema.UsageVariableType

	if valNode.ShortTag() == "!!map" {
		usageValueType = schema.Items

		if len(valNode.Content)%2 != 0 {
			log.Errorf("YAML map node contents are not divisible by 2")
			return nil, errors.New("unexpected YAML format")
		}

		items := make([]*schema.UsageItem, 0, len(valNode.Content)/2)

		for i := 0; i < len(valNode.Content); i += 2 {
			mapKeyNode := valNode.Content[i]
			mapValNode := valNode.Content[i+1]

			mapUsageItem, err := yamlToUsageItem(mapKeyNode, mapValNode)
			if err != nil {
				return nil, err
			}

			items = append(items, mapUsageItem)
		}

		value = items
	} else {
		err := valNode.Decode(&value)
		if err != nil {
			log.Errorf("Unable to decode YAML value")
			return nil, errors.New("unexpected YAML format")
		}

		switch valNode.ShortTag() {
		case "!!int":
			usageValueType = schema.Int64

		case "!!float":
			usageValueType = schema.Float64

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

func resourceUsagesToYAML(resourceUsages []*ResourceUsage) (yamlv3.Node, bool) {
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

			rawValue := item.DefaultValue
			itemNodeIsCommented := true

			if item.Value != nil {
				rawValue = item.Value
				itemNodeIsCommented = false
				resourceNodeIsCommented = false
				rootNodeIsCommented = false
			}

			if item.ValueType == schema.Items {
				if rawValue != nil {
					subResourceItems := rawValue.([]*schema.UsageItem)
					subResourceUsage := &ResourceUsage{
						Name:  item.Key,
						Items: subResourceItems,
					}
					subResourceValNode, _ := resourceUsagesToYAML([]*ResourceUsage{subResourceUsage})
					resourceValNode.Content = append(resourceValNode.Content, subResourceValNode.Content...)
				}
				continue
			}

			var tag string
			var value string

			switch item.ValueType {
			case schema.Float64:
				tag = "!!float"
				value = fmt.Sprintf("%f", rawValue)
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
			case schema.Items:
				tag = "!!map"
				kind = yamlv3.MappingNode
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

// markNodeAsComment marks a node as a comment which we then post process later to add the #
// We could use the yamlv3 FootComment/LineComment but this gets complicated with indentation
// especially when we have edge cases like resources that are fully commented out
func markNodeAsComment(node *yamlv3.Node) {
	node.Value = commentMark + node.Value
}

func replaceCommentMarks(b []byte, resourceUsageNodeIsCommented bool) []byte {
	if resourceUsageNodeIsCommented {
		b = bytes.Replace(b, []byte("resource_usage:"), []byte("# resource_usage:"), 1)
	}

	return bytes.ReplaceAll(b, []byte(commentMark), []byte("# "))
}
