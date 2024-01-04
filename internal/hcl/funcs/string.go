package funcs

import (
	"regexp"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// EndsWithFunc constructs a function that checks if a string ends with suffix.
var EndsWithFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
		{
			Name: "suffix",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		str := args[0].AsString()
		suffix := args[1].AsString()

		return cty.BoolVal(strings.HasSuffix(str, suffix)), nil
	},
})

// EndsWith checks if a string ends with suffix.
func EndsWith(str, suffix cty.Value) (cty.Value, error) {
	return EndsWithFunc.Call([]cty.Value{str, suffix})
}

// ReplaceFunc constructs a function that searches a given string for another
// given substring, and replaces each occurrence with a given replacement string.
var ReplaceFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
		{
			Name: "substr",
			Type: cty.String,
		},
		{
			Name: "replace",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		str := args[0].AsString()
		substr := args[1].AsString()
		replace := args[2].AsString()

		// We search/replace using a regexp if the string is surrounded
		// in forward slashes.
		if len(substr) > 1 && substr[0] == '/' && substr[len(substr)-1] == '/' {
			re, err := regexp.Compile(substr[1 : len(substr)-1])
			if err != nil {
				return cty.UnknownVal(cty.String), err
			}

			return cty.StringVal(re.ReplaceAllString(str, replace)), nil
		}

		return cty.StringVal(strings.ReplaceAll(str, substr, replace)), nil
	},
})

// Replace searches a given string for another given substring,
// and replaces all occurrences with a given replacement string.
func Replace(str, substr, replace cty.Value) (cty.Value, error) {
	return ReplaceFunc.Call([]cty.Value{str, substr, replace})
}

// StartsWithFunc constructs a function that checks if a string begins with prefix.
var StartsWithFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
		{
			Name: "prefix",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		str := args[0].AsString()
		prefix := args[1].AsString()

		return cty.BoolVal(strings.HasPrefix(str, prefix)), nil
	},
})

// StartsWith checks if a string begins with prefix.
func StartsWith(str, prefix cty.Value) (cty.Value, error) {
	return StartsWithFunc.Call([]cty.Value{str, prefix})
}

// StrContainsFunc constructs a function that checks if a string contains substr.
var StrContainsFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
		{
			Name: "substr",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		str := args[0].AsString()
		substr := args[1].AsString()

		return cty.BoolVal(strings.Contains(str, substr)), nil
	},
})

// StrContains checks if a string contains substr.
func StrContains(str, substr cty.Value) (cty.Value, error) {
	return StrContainsFunc.Call([]cty.Value{str, substr})
}
