package funcs

import (
	"errors"
	"fmt"
	"maps"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

// CoalesceFunc constructs a function that takes any number of arguments and
// returns the first one that isn't empty. This function was copied from go-cty
// stdlib and modified so that it returns the first *non-empty* non-null element
// from a sequence, instead of merely the first non-null.
var CoalesceFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "vals",
		Type:             cty.DynamicPseudoType,
		AllowUnknown:     true,
		AllowDynamicType: true,
		AllowNull:        true,
	},
	Type: func(args []cty.Value) (ret cty.Type, err error) {
		argTypes := make([]cty.Type, len(args))
		for i, val := range args {
			argTypes[i] = val.Type()
		}
		retType, _ := convert.UnifyUnsafe(argTypes)
		if retType == cty.NilType {
			return cty.NilType, errors.New("all arguments must have the same type")
		}
		return retType, nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		for _, argVal := range args {
			// We already know this will succeed because of the checks in our Type func above
			argVal, _ = convert.Convert(argVal, retType)
			if !argVal.IsKnown() {
				return cty.UnknownVal(retType), nil
			}
			if argVal.IsNull() {
				continue
			}
			if retType == cty.String && argVal.RawEquals(cty.StringVal("")) {
				continue
			}

			return argVal, nil
		}
		return cty.NilVal, errors.New("no non-null, non-empty-string arguments")
	},
})

// MergeFunc is an Infracost specific version of collection.MergeFunc which
// handles Infracost mocked return values. If the argument contains an Infracost mock
// string then we ignore it in the merge.
var MergeFunc = function.New(&function.Spec{
	Description: `Merges all of the elements from the given maps into a single map, or the attributes from given objects into a single object.`,
	Params:      []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "maps",
		Type:             cty.DynamicPseudoType,
		AllowUnknown:     true,
		AllowDynamicType: true,
		AllowNull:        true,
		AllowMarked:      true,
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		// empty args is accepted, so assume an empty object since we have no
		// key-value types.
		if len(args) == 0 {
			return cty.EmptyObject, nil
		}

		// collect the possible object attrs
		attrs := map[string]cty.Type{}

		first := cty.NilType
		matching := true
		attrsKnown := true
		for i, arg := range args {
			ty := arg.Type()
			// marks are attached to values, so ignore while determining type
			arg, _ = arg.Unmark()

			switch {
			case ty.IsObjectType() && !arg.IsNull():
				maps.Copy(attrs, ty.AttributeTypes())
			case ty.IsMapType():
				switch {
				case arg.IsNull():
					// pass, nothing to add
				case arg.IsKnown():
					ety := arg.Type().ElementType()
					for it := arg.ElementIterator(); it.Next(); {
						attr, _ := it.Element()
						attrs[attr.AsString()] = ety
					}
				default:
					// any non-object/map values will get here
					// any unknown maps means we don't know all possible attrs
					// for the return type
					attrsKnown = false
				}
			}

			// record the first argument type for comparison
			if i == 0 {
				first = arg.Type()
				continue
			}

			if !ty.Equals(first) && matching {
				matching = false
			}
		}

		// the types all match, so use the first argument type
		if matching {
			return first, nil
		}

		// We had a mix of unknown maps and objects, so we can't predict the
		// attributes
		if !attrsKnown {
			return cty.DynamicPseudoType, nil
		}

		return cty.Object(attrs), nil
	},
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		outputMap := make(map[string]cty.Value)
		var markses []cty.ValueMarks // remember any marked maps/objects we find

		for _, arg := range args {
			// We skip uninterable values because we might get mock values here
			if arg.IsNull() || !arg.CanIterateElements() {
				continue
			}
			arg, argMarks := arg.Unmark()
			if len(argMarks) > 0 {
				markses = append(markses, argMarks)
			}
			for it := arg.ElementIterator(); it.Next(); {
				k, v := it.Element()
				outputMap[k.AsString()] = v
			}
		}

		switch {
		case retType.IsMapType():
			if len(outputMap) == 0 {
				return cty.MapValEmpty(retType.ElementType()).WithMarks(markses...), nil
			}
			return cty.MapVal(outputMap).WithMarks(markses...), nil
		case retType.IsObjectType(), retType.Equals(cty.DynamicPseudoType):
			return cty.ObjectVal(outputMap).WithMarks(markses...), nil
		default:
			panic(fmt.Sprintf("unexpected return type: %#v", retType))
		}
	},
})

// Coalesce takes any number of arguments and returns the first one that isn't empty.
func Coalesce(args ...cty.Value) (cty.Value, error) {
	return CoalesceFunc.Call(args)
}
