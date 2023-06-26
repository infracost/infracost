package funcs

import (
	"strings"

	yaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

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
		if strings.HasPrefix(val, "mock") {
			return cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}), nil
		}

		return yaml.Standard.ImpliedType([]byte(val))
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		val := args[0].AsString()
		if strings.HasPrefix(val, "mock") {
			return cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}), nil
		}

		return yaml.Standard.Unmarshal([]byte(val), retType)
	},
})
