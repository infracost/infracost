package funcs

import (
	"strings"

	"github.com/infracost/infracost/internal/hcl/mock"
	yaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// YAMLDecodeFunc is an Infracost specific version of the yaml.YAMLDecodeFunc
// which handles Infracost mocked return values. If the argument passed to YAMLDecodeFunc
// is an Infracost mock (e.g. a string with value mock-value) then we return a mocked object
// that can be used in the HCL evaluation loop. This means we get less unwanted nil values when
// evaluating HCL files. This is especially important when evaluating Terragrunt HCL files
// as unexpected nils cause program termination.
var YAMLDecodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "src",
			Type: cty.String,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		if !args[0].IsKnown() {
			return cty.DynamicPseudoType, nil
		}
		if args[0].IsNull() {
			return cty.NilType, function.NewArgErrorf(0, "YAML source code cannot be null")
		}
		val := args[0].AsString()
		if strings.HasSuffix(val, "-mock") || strings.Contains(val, mock.Identifier) {
			return cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}), nil
		}

		return yaml.Standard.ImpliedType([]byte(val))
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		val := args[0].AsString()
		if strings.HasSuffix(val, "-mock") || strings.Contains(val, mock.Identifier) {
			return cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}), nil
		}

		return yaml.Standard.Unmarshal([]byte(val), retType)
	},
})
