package usage

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/infracost/infracost/internal/metrics"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

const minUsageFileVersion = "0.1"
const maxUsageFileVersion = "0.1"

type UsageFile struct { // nolint:revive
	Version string `yaml:"version"`
	// We represent resource type usage in using a YAML node so we have control over the comments
	RawResourceTypeUsage yamlv3.Node `yaml:"resource_type_default_usage"`
	// The raw usage is then parsed into this struct
	ResourceTypeUsages []*ResourceUsage `yaml:"-"`
	// We represent resource usage in using a YAML node so we have control over the comments
	RawResourceUsage yamlv3.Node `yaml:"resource_usage"`
	// The raw usage is then parsed into this struct
	ResourceUsages []*ResourceUsage `yaml:"-"`
}

// CreateUsageFile creates a blank usage file if it does not exists
func CreateUsageFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		usageFile := NewBlankUsageFile()

		err = usageFile.WriteToPath(path)
		if err != nil {
			return errors.Wrapf(err, "Error writing blank usage file to %s", path)
		}
	} else {
		logging.Logger.Debug().Msg("Specified usage file already exists, no overriding")
	}

	return nil
}

func LoadUsageFile(path string) (*UsageFile, error) {
	loadUsageTimer := metrics.GetTimer("parallel_runner.load_usage.duration", false, path).Start()
	defer loadUsageTimer.Stop()
	blankUsage := NewBlankUsageFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logging.Logger.Debug().Msg("Specified usage file does not exist. Using a blank file")

		return blankUsage, nil
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return blankUsage, errors.Wrapf(err, "Error reading usage file")
	}

	usageFile, err := LoadUsageFileFromString(string(contents))
	if err != nil {
		return blankUsage, errors.Wrapf(err, "Error loading usage file")
	}

	return usageFile, nil
}

func NewBlankUsageFile() *UsageFile {
	usageFile := &UsageFile{
		Version: maxUsageFileVersion,
		RawResourceTypeUsage: yamlv3.Node{
			Kind: yamlv3.MappingNode,
		},
		RawResourceUsage: yamlv3.Node{
			Kind: yamlv3.MappingNode,
		},
	}

	return usageFile
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

	err = usageFile.parseResourceUsages()
	if err != nil {
		return usageFile, errors.Wrap(err, "Error loading YAML file")
	}

	return usageFile, nil
}

func (u *UsageFile) WriteToPath(path string) error {
	allResourceTypesCommented, allResourcesCommented := u.dumpResourceUsages()

	root := &yamlv3.Node{
		Kind: yamlv3.MappingNode,
		HeadComment: `You can use this file to define resource usage estimates for Infracost to use when calculating
the cost of usage-based resource, such as AWS S3 or Lambda.
` + "`infracost breakdown --usage-file infracost-usage.yml [other flags]`" + `
See https://infracost.io/usage-file/ for docs`,
	}

	resourceTypeUsagesKeyNode := &yamlv3.Node{
		Kind:  yamlv3.ScalarNode,
		Value: "resource_type_default_usage",
	}
	resourceUsagesKeyNode := &yamlv3.Node{
		Kind:  yamlv3.ScalarNode,
		Value: "resource_usage",
	}

	if allResourceTypesCommented {
		markNodeAsComment(resourceTypeUsagesKeyNode)
	}
	if allResourcesCommented {
		markNodeAsComment(resourceUsagesKeyNode)
	}

	root.Content = append(root.Content,
		&yamlv3.Node{
			Kind:  yamlv3.ScalarNode,
			Value: "version",
		},
		&yamlv3.Node{
			Kind:  yamlv3.ScalarNode,
			Value: u.Version,
		},
		resourceTypeUsagesKeyNode,
		&u.RawResourceTypeUsage,
		resourceUsagesKeyNode,
		&u.RawResourceUsage,
	)

	// Add a comment to the first commented-out resource
	for _, node := range u.RawResourceTypeUsage.Content {
		if isNodeMarkedAsCommented(node) {
			node.HeadComment = `##
## The following usage values apply to each resource of the given type, which is useful when you want to define defaults.
## All values are commented-out, you can uncomment resource types and customize as needed.
##`
			break
		}
	}

	// Add a comment to the first commented-out resource
	for _, node := range u.RawResourceUsage.Content {
		if isNodeMarkedAsCommented(node) {
			node.HeadComment = `##
## The following usage values apply to individual resources and override any value defined in the resource_type_default_usage section.
## All values are commented-out, you can uncomment resources and customize as needed.
##`
			break
		}
	}

	var buf bytes.Buffer
	yamlEncoder := yamlv3.NewEncoder(&buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(root)
	if err != nil {
		return err
	}

	b := buf.Bytes()
	// If all the resources are commented then we also want to comment the resource_usage key
	b = replaceCommentMarks(b)

	return os.WriteFile(path, b, 0600)
}

func (u *UsageFile) ToUsageDataMap() schema.UsageMap {
	m := make(map[string]any)

	for _, resourceUsage := range u.ResourceTypeUsages {
		m[resourceUsage.Name] = resourceUsage.Map()
	}

	for _, resourceUsage := range u.ResourceUsages {
		m[resourceUsage.Name] = resourceUsage.Map()
	}

	return schema.NewUsageMapFromInterface(m)
}

func (u *UsageFile) checkVersion() bool {
	v := u.Version
	if !strings.HasPrefix(u.Version, "v") {
		v = "v" + u.Version
	}
	return semver.Compare(v, "v"+minUsageFileVersion) >= 0 && semver.Compare(v, "v"+maxUsageFileVersion) <= 0
}

// InvalidKeys returns a list of keys that are invalid in the usage file.
// It currently checks the reference usage file for a list of valid keys.
// In the future we will want this to usage the resource usage schema structs as well.
func (u *UsageFile) InvalidKeys() ([]string, error) {
	invalidKeys := make([]string, 0)

	refFile, err := LoadReferenceFile()
	if err != nil {
		return invalidKeys, err
	}

	for _, resourceUsage := range u.ResourceUsages {
		refResourceUsage := refFile.FindMatchingResourceUsage(resourceUsage.Name)
		if refResourceUsage == nil {
			continue
		}
		refItemMap := refResourceUsage.Map()

		// Iterate over provided keys and check if they are
		// present in the reference usage file
		for _, item := range resourceUsage.Items {
			invalidKeys = append(invalidKeys, findInvalidKeys(item, refItemMap)...)
		}
	}

	for _, resourceUsage := range u.ResourceTypeUsages {
		refResourceUsage := refFile.FindMatchingResourceTypeUsage(resourceUsage.Name)
		if refResourceUsage == nil {
			continue
		}

		refItemMap := refResourceUsage.Map()

		// Iterate over provided keys and check if they are
		// present in the reference usage file
		for _, item := range resourceUsage.Items {
			invalidKeys = append(invalidKeys, findInvalidKeys(item, refItemMap)...)
		}
	}

	// Remove duplicate entries
	invalidKeys = removeDuplicateStr(invalidKeys)

	// Sort the keys alphabetically
	sort.Strings(invalidKeys)

	return invalidKeys, nil
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// findInvalidKeys recursively searches for invalid keys in the provided item
func findInvalidKeys(item *schema.UsageItem, refMap map[string]any) []string {
	invalidKeys := make([]string, 0)

	if refVal, ok := refMap[item.Key]; !ok {
		invalidKeys = append(invalidKeys, item.Key)
	} else if item.ValueType == schema.SubResourceUsage && item.Value != nil {
		for _, subItem := range item.Value.(*ResourceUsage).Items {
			invalidKeys = append(invalidKeys, findInvalidKeys(subItem, refVal.(map[string]any))...)
		}
	}

	return invalidKeys
}

func (u *UsageFile) parseResourceUsages() error {
	var err error
	u.ResourceUsages, err = ResourceUsagesFromYAML(u.RawResourceUsage)
	if err != nil {
		return errors.Wrapf(err, "Error parsing usage file")
	}
	u.ResourceTypeUsages, err = ResourceUsagesFromYAML(u.RawResourceTypeUsage)
	if err != nil {
		return errors.Wrapf(err, "Error parsing usage file")
	}
	return nil
}

func (u *UsageFile) dumpResourceUsages() (bool, bool) {
	var allResourceTypesCommented bool
	var allResourcesCommented bool
	u.RawResourceTypeUsage, allResourceTypesCommented = ResourceUsagesToYAML(u.ResourceTypeUsages)
	u.RawResourceUsage, allResourcesCommented = ResourceUsagesToYAML(u.ResourceUsages)
	return allResourceTypesCommented, allResourcesCommented
}
