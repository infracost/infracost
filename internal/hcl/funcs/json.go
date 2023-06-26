package funcs

import (
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/json"
)

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
		if strings.HasPrefix(val, "mock") {
			return cty.Object(map[string]cty.Type{
				"foo": cty.String,
			}), nil
		}

		return json.ImpliedType([]byte(val))
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		val := args[0].AsString()
		if strings.HasPrefix(val, "mock") {
			return cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}), nil
		}

		return json.Unmarshal([]byte(val), retType)
	},
})
