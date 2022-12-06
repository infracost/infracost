package funcs

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// PrintArgs prints any number of args to the std out using fmt.
// This can be used for debugging variables/inputs at points in the Terraform evaluation.
// Example usage:
//
//	infracostprint("test", 50)
//
// will print:
//
//	"terraform print cty.StringVal("test"), cty.IntVal(50)
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
			Name: "path",
			Type: cty.String,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		return args[0].Type(), nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		var strs = make([]string, len(args))
		for i, arg := range args {
			strs[i] = arg.GoString()
		}

		list := strings.Join(strs, ", ")
		fmt.Printf("terraform print %s\n", list)

		return args[0], nil
	},
})

// LogArgs is identical to PrintArgs but writes the arguments to the Infracost log.
// This is useful to understand arguments as they change in the module evaluation.
// As the arguments will be printed next to log entries that correspond to the program runtime.
// e.g:
//
//	root_block_device {
//		volume_size = infracostlog("test")
//	}
//
// will log:
//
//	time="2022-12-06T10:27:40Z" level=debug enable_cloud_org=false ... attribute_name=volume_size provider=terraform_dir block_name=root_block_device. sync_usage=false msg="fetching attribute value"
//	time="2022-12-06T10:27:40Z" level=debug ... msg="terraform print: cty.StringVal(\"test\")"
func LogArgs(logger *logrus.Entry) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			return args[0].Type(), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			var strs = make([]string, len(args))
			for i, arg := range args {
				strs[i] = arg.GoString()
			}

			list := strings.Join(strs, ", ")
			logger.Logger.Debugf("terraform print: %s", list)

			return args[0], nil
		},
	})
}
