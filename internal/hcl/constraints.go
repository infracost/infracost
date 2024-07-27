package hcl

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/zclconf/go-cty/cty/gocty"
	"strings"
)

var (
	// AWSVersionConstraintVolumeTags is the version constraint for the Terraform aws
	// provider that supports propagating default tags to volume tags. See
	// https://github.com/hashicorp/terraform-provider-aws/issues/19890 for more
	// information.
	AWSVersionConstraintVolumeTags = version.Must(version.NewVersion("5.39.0"))
)

// ProviderConstraints represents the constraints for a providers within a
// module. This can be used to check if Infracost functionality is applicable to
// a given run.
type ProviderConstraints struct {
	AWS    version.Constraints
	Google version.Constraints
}

const constraintOpChars = "=><!~"

func splitConstraint(c string) (string, string) {
	c = strings.TrimSpace(c)
	ver := c
	var op string
	for i, r := range c {
		if !strings.ContainsRune(constraintOpChars, r) {
			break
		}
		op = c[:i+1]
		ver = c[i+1:]
	}
	return op, strings.TrimSpace(ver)
}

// ConstraintsAllowVersionOrAbove checks if the given constraints enforce at least the given minVersion.
func ConstraintsAllowVersionOrAbove(constraints version.Constraints, requiredVersion *version.Version) bool {

	for _, c := range constraints {

		op, ver := splitConstraint(c.String())
		check, err := version.NewVersion(ver)
		if err != nil {
			continue
		}
		switch op {
		case "~>":
			segments := check.Segments()
			if len(segments) < 2 {
				continue
			}
			requiredSegments := requiredVersion.Segments()
			if len(requiredSegments) < 2 {
				continue
			}
			segCount := strings.Count(ver, ".") + 1
			if segCount > 1 {
				if segments[0] < requiredSegments[0] {
					return false
				}
				// only check the minor version if the major version is the same AND the path version is the rightmost
				// specified component in the ~> constraint
				if segCount > 2 {
					if segments[0] == requiredSegments[0] && segments[1] < requiredSegments[1] {
						return false
					}
				}
			}
		case "<":
			if check.LessThanOrEqual(requiredVersion) {
				return false
			}
		case "<=":
			if check.LessThan(requiredVersion) {
				return false
			}
		case "", "=":
			if check.LessThan(requiredVersion) {
				return false
			}
		}
	}

	return true
}

// UnmarshalJSON parses the JSON representation of the ProviderConstraints and
// sets the constraints for the sub providers.
func (p *ProviderConstraints) UnmarshalJSON(b []byte) error {
	var out map[string]string
	err := json.Unmarshal(b, &out)
	if err != nil {
		return err
	}

	if v, ok := out["aws"]; ok {
		constraints, err := version.NewConstraint(v)
		if err == nil {
			p.AWS = constraints
		}
	}

	return nil
}

// MarshalJSON returns the JSON representation of the ProviderConstraints.
// This is used to serialize the constraints for the sub providers.
func (p *ProviderConstraints) MarshalJSON() ([]byte, error) {
	out := map[string]string{}
	if p == nil {
		return json.Marshal(&out)
	}

	if p.AWS != nil {
		out["aws"] = p.AWS.String()
	} else {
		out["aws"] = ""
	}

	return json.Marshal(out)
}

// NewProviderConstraints parses the provider blocks for any Terraform
// configuration blocks if found it will attempt to return a ProviderConstraints
// struct from the required_providers configuration. Currently, we only support
// AWS provider constraints.
func NewProviderConstraints(blocks Blocks) *ProviderConstraints {
	var allConstraints ProviderConstraints
	terraformBlocks := blocks.OfType("terraform")
	for _, block := range terraformBlocks {
		req := block.GetChildBlock("required_providers")
		if req == nil {
			continue
		}

		if constraints, err := readProviderConstraints(req, "aws"); err == nil {
			allConstraints.AWS = constraints
		}
		if constraints, err := readProviderConstraints(req, "google"); err == nil {
			allConstraints.Google = constraints
		}
	}
	return &allConstraints
}

func readProviderConstraints(requiredProviderBlock *Block, provider string) (version.Constraints, error) {
	def := requiredProviderBlock.Values().AsValueMap()

	body, ok := def[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", provider)
	}

	if body.IsNull() || !body.IsKnown() {
		return nil, fmt.Errorf("provider %s is not a known type", provider)
	}

	var source string
	var constraintVersion string
	if body.Type().IsObjectType() {
		// v0.13 required provider definition
		constraintDef := body.AsValueMap()
		_ = gocty.FromCtyValue(constraintDef["source"], &source)
		if source != fmt.Sprintf("hashicorp/%s", provider) {
			return nil, fmt.Errorf("provider %s has an unsupported source: %s", provider, source)
		}

		_ = gocty.FromCtyValue(constraintDef["version"], &constraintVersion)
	} else {
		// support v0.12 provider definition
		// https://developer.hashicorp.com/terraform/language/providers/requirements#v0-12-compatible-provider-requirements
		_ = gocty.FromCtyValue(body, &constraintVersion)
	}

	// parse the version and return it
	return version.NewConstraint(constraintVersion)
}
