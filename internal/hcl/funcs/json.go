package funcs

import (
	"strings"

	"github.com/infracost/infracost/internal/hcl/mock"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/json"
)

// JSONDecodeFunc is an Infracost specific version of the json.JSONDecodeFunc
// which handles Infracost mocked return values. If the argument passed to JSONDecodeFunc
// is an Infracost mock (e.g. a string with value mock-value) then we return a mocked object
// that can be used in the HCL evaluation loop. This means we get less unwanted nil values when
// evaluating HCL files. This is especially important when evaluating Terragrunt HCL files
// as unexpected nils cause program termination.
var JSONDecodeFunc = function.New(&function.Spec{
	Description: `Parses the given string as JSON and returns a value corresponding to what the JSON document describes.`,
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		str := args[0]
		if !str.IsKnown() {
			return cty.DynamicPseudoType, nil
		}

		val := str.AsString()
		if strings.HasSuffix(val, "-mock") || strings.Contains(val, mock.Identifier) {
			return cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}), nil
		}

		return json.ImpliedType([]byte(val))
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		val := args[0].AsString()
		if strings.HasSuffix(val, "-mock") || strings.Contains(val, mock.Identifier) {
			return cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}), nil
		}

		return json.Unmarshal([]byte(val), retType)
	},
})
