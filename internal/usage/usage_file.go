package usage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	yamlv3 "gopkg.in/yaml.v3"
)

const minUsageFileVersion = "0.1"
const maxUsageFileVersion = "0.1"

type UsageFile struct { // nolint:revive
	Version string `yaml:"version"`
	// We represent resource usage in using a YAML node so we have control over the comments
	RawResourceUsage yamlv3.Node `yaml:"resource_usage"`
	// The raw usage is then parsed into this struct
	ResourceUsages []*ResourceUsage `yaml:"-"`
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

	err = usageFile.parseResourceUsages()
	if err != nil {
		return usageFile, errors.Wrap(err, "Error loading YAML file")
	}

	return usageFile, nil
}

func (u *UsageFile) WriteToPath(path string) error {
	allCommented := u.dumpResourceUsages()

	var buf bytes.Buffer
	yamlEncoder := yamlv3.NewEncoder(&buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(u)
	if err != nil {
		return err
	}

	b := buf.Bytes()

	// If all the resources are commented then we also want to comment the resource_usage key
	if allCommented {
		b = bytes.Replace(b, []byte("resource_usage:"), []byte("# resource_usage:"), 1)
	}
	b = replaceCommentMarks(b)

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
func findInvalidKeys(item *schema.UsageItem, refMap map[string]interface{}) []string {
	invalidKeys := make([]string, 0)

	if refVal, ok := refMap[item.Key]; !ok {
		invalidKeys = append(invalidKeys, item.Key)
	} else if item.ValueType == schema.SubResourceUsage && item.Value != nil {
		for _, subItem := range item.Value.(*ResourceUsage).Items {
			invalidKeys = append(invalidKeys, findInvalidKeys(subItem, refVal.(map[string]interface{}))...)
		}
	}

	return invalidKeys
}

func (u *UsageFile) parseResourceUsages() error {
	var err error
	u.ResourceUsages, err = ResourceUsagesFromYAML(u.RawResourceUsage)
	return err
}

func (u *UsageFile) dumpResourceUsages() bool {
	var allCommented bool
	u.RawResourceUsage, allCommented = ResourceUsagesToYAML(u.ResourceUsages)
	return allCommented
}
