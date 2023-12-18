package funcs

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	ctyJson "github.com/zclconf/go-cty/cty/json"
)

// PrintArgs prints any number of args to the std out using fmt.
// This can be used for debugging variables/inputs at points in the Terraform evaluation.
// Example usage:
//
//	infracostprint("test", 50)
//
// will print:
//
//	"terraform print "test":cty.IntVal(50)
//
// PrintArgs will return any args passed unaltered so that the args are still safe to use in the evaluation context.
// e.g:
//
//	locals {
//		test = infracostprint("a")
//	}
//
// will still have `local.test` == "a" if used by other Terraform attributes/blocks. This allows debugging to unalter the
// Terraform evaluation and not cause unwanted consequences.
var PrintArgs = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "name",
			Type: cty.String,
		},
		{
			Name:             "v",
			Type:             cty.DynamicPseudoType,
			AllowNull:        true,
			AllowUnknown:     true,
			AllowMarked:      true,
			AllowDynamicType: true,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		return args[1].Type(), nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		fmt.Printf("terraform print %q:%s\n", args[0].AsString(), string(valueToBytes(args[1])))

		return args[1], nil
	},
})

func valueToBytes(v cty.Value) []byte {
	simple := ctyJson.SimpleJSONValue{Value: v}
	b, _ := simple.MarshalJSON()
	return b
}

// LogArgs is identical to PrintArgs but writes the arguments to the Infracost log.
// This is useful to understand arguments as they change in the module evaluation.
// As the arguments will be printed next to log entries that correspond to the program runtime.
// e.g:
//
//	root_block_device {
//		volume_size = infracostlog("test", "foo")
//	}
//
// will log:
//
//	time="2022-12-06T10:27:40Z" level=debug enable_cloud_org=false ... attribute_name=volume_size provider=terraform_dir block_name=root_block_device. sync_usage=false msg="fetching attribute value"
//	time="2022-12-06T10:27:40Z" level=debug ... msg="terraform print "test":cty.StringVal(\"foo\")"
func LogArgs(logger zerolog.Logger) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "name",
				Type: cty.String,
			},
			{
				Name:             "v",
				Type:             cty.DynamicPseudoType,
				AllowNull:        true,
				AllowMarked:      true,
				AllowUnknown:     true,
				AllowDynamicType: true,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			return args[1].Type(), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			logger.Debug().Msgf("terraform print %q:%s", args[0], args[1].GoString())

			return args[1], nil
		},
	})
}
