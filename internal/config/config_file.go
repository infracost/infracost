package config

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

const (
	minConfigFileVersion = "0.1"
	maxConfigFileVersion = "0.1"
)

var ErrorInvalidConfigFile = errors.New("parsing config file failed, check file syntax")

// YamlError is a custom error type that allows setting multiple
// error messages under a base message. It is used to decipher
// between internal errors and the yaml.v2 errors.
type YamlError struct {
	// raw can be used to override the base & errors formatting
	// and just use a single error value.
	raw error

	base   string
	errors []error
	indent int
}

func (y *YamlError) add(err error) {
	y.errors = append(y.errors, err)
}

func (y *YamlError) isValid() bool {
	if y.raw != nil {
		return true
	}

	return len(y.errors) > 0
}

// Error implements the error interface returning the YamlError as a string.
// If a raw error is set it simply returns the error message from that.
// Otherwise, it constructs an indented error message out of the base and errors.
//
// YamlError.Error supports multiple nesting and can construct heavily indented output if needed.
// e.g.
//
//	&YamlError{
//		base: "top message",
//		errors: []error{
//			errors.New("top error 1"),
//			&YamlError{
//				base: "child message",
//				errors: []error{
//					errors.New("child error 1"),
//				},
//			},
//		},
//	}
//
// would output a string like so:
//
//	top message:
//		top error 1
//		child message:
//			child error 1
//
// This can be useful for ui error messages where you need to highlight issues
// with specific fields/entries.
func (y *YamlError) Error() string {
	if y.raw != nil {
		return y.raw.Error()
	}

	if y.indent == 0 {
		y.indent = 1
	}

	indent := "\t"
	if y.indent > 1 {
		indent = strings.Repeat(indent, y.indent)
	}

	str := y.base + ":\n"

	for i, err := range y.errors {
		if v, ok := err.(*YamlError); ok {
			v.indent = y.indent + 1
		}

		if i == len(y.errors)-1 {
			str += indent + err.Error()
			break
		}

		str += indent + err.Error() + "\n"
	}

	return str
}

type ConfigFileSpec struct {
	Version                 string                  `yaml:"version"`
	Projects                []*Project              `yaml:"projects" ignored:"true"`
	TerraformSourceMapRegex TerraformSourceMapRegex `yaml:"terraform_source_map,omitempty"`
}

// UnmarshalYAML implements the yaml.v2.Unmarshaller interface. Marshalls the
// yaml into an intermediary struct so that we can catch field violations before
// the data is set on the main ConfigFileSpec. Note this method must return a YamlError
// type so that we don't run into error collisions with the base yaml.v2 errors.
func (f *ConfigFileSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type roughFile struct {
		Version                 string                   `yaml:"version"`
		Projects                []map[string]interface{} `yaml:"projects"`
		TerraformSourceMapRegex TerraformSourceMapRegex  `yaml:"terraform_source_map,omitempty"`
	}

	var r roughFile
	err := unmarshal(&r)
	if err != nil {
		return &YamlError{raw: ErrorInvalidConfigFile}
	}

	p := Project{}
	v := reflect.TypeOf(p)
	numFields := v.NumField()
	allowedKeys := make(map[string]struct{}, numFields)

	for i := 0; i < numFields; i++ {
		tag := v.Field(i).Tag.Get("yaml")
		pieces := strings.Split(tag, ",")
		allowedKeys[strings.TrimSpace(pieces[0])] = struct{}{}
	}

	validationError := &YamlError{
		base: "config file is invalid, see https://infracost.io/config-file for valid options",
	}

	for i, fields := range r.Projects {
		if v, ok := fields["path"]; !ok || v == nil {
			validationError.add(&YamlError{
				base:   fmt.Sprintf("project config at index %d was invalid", i),
				errors: []error{errors.New("project must have a valid path definition")},
			})
			continue
		}

		projectError := &YamlError{
			base: fmt.Sprintf("project config defined for path: [%s] is invalid", fields["path"]),
		}

		sorted := make([]string, 0, len(fields))
		for k := range fields {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)

		for _, k := range sorted {
			if _, ok := allowedKeys[k]; ok {
				continue
			}

			projectError.add(fmt.Errorf("%s is not a valid project configuration option", k))
		}

		if projectError.isValid() {
			validationError.add(projectError)
		}
	}

	if validationError.isValid() {
		return validationError
	}

	if !checkVersion(r.Version) {
		return &YamlError{
			base: "config file is invalid, see https://infracost.io/config-file for file specification",
			errors: []error{
				fmt.Errorf("version '%s' is not supported, valid versions are %s ≤ x ≤ %s", r.Version, minConfigFileVersion, maxConfigFileVersion),
			},
			indent: 0,
		}
	}

	type fileSpecClone ConfigFileSpec
	var c fileSpecClone
	err = unmarshal(&c)
	if err != nil {
		return &YamlError{raw: ErrorInvalidConfigFile}
	}

	f.Version = c.Version
	f.Projects = c.Projects
	f.TerraformSourceMapRegex = c.TerraformSourceMapRegex
	return nil
}

func LoadConfigFile(path string) (ConfigFileSpec, error) {
	var cfgFile ConfigFileSpec

	if !FileExists(path) {
		return cfgFile, fmt.Errorf("config file does not exist at %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return cfgFile, fmt.Errorf("%w: %s", ErrorInvalidConfigFile, err)
	}

	content = []byte(os.ExpandEnv(string(content)))

	err = yaml.Unmarshal(content, &cfgFile)
	if err != nil {
		// we have to make this custom error type checking here
		// as indentations cause the yaml.Unmarshal to panic
		// it catches the panic and returns an error but in order
		// not to stutter the errors we should check here for
		// our custom error type.
		if _, ok := err.(*YamlError); ok {
			return cfgFile, err
		}

		// if we receive a caught panic error, wrap the message in something more user-friendly
		return cfgFile, fmt.Errorf("%w: %s", ErrorInvalidConfigFile, err)
	}

	return cfgFile, nil
}

func checkVersion(v string) bool {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return semver.Compare(v, "v"+minConfigFileVersion) >= 0 && semver.Compare(v, "v"+maxConfigFileVersion) <= 0
}
