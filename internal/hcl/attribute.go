package hcl

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/hcl/mock"
)

var (
	missingAttributeDiagnostic        = "Unsupported attribute"
	valueIsNonIterableDiagnostic      = "Iteration over non-iterable value"
	invalidFunctionArgumentDiagnostic = "Invalid function argument"
	unknownVariableDiagnostic         = "Unknown variable"
)

// Attribute provides a wrapper struct around hcl.Attribute it provides
// helper methods and functionality for common interactions with hcl.Attribute.
//
// Attributes are key/value pairs that are part of a Block. For example take the following Block:
//
//			resource "aws_instance" "t3_standard" {
//			  	ami           = "fake_ami"
//	 		instance_type = "t3.medium"
//
//	 		credit_specification {
//	   			cpu_credits = "standard"
//	 		}
//			}
//
// "ami" & "instance_type" are the Attributes of this Block, "credit_specification" is a child Block
// see Block.Children for more info.
type Attribute struct {
	// HCLAttr is the underlying hcl.Attribute that the Attribute references.
	HCLAttr *hcl.Attribute
	// Ctx is the context that the Attribute should be evaluated against. This propagates
	// any references from variables into the attribute.
	Ctx *Context
	// Verbose defines if the attribute should log verbose diagnostics messages to debug.
	Verbose bool
	Logger  zerolog.Logger
	// IsGraph is a flag that indicates if the attribute should be evaluated with the graph evaluation
	IsGraph bool
	// newMock generates a mock value for the attribute if it's value is missing.
	newMock                func(attr *Attribute) cty.Value
	previousValue          cty.Value
	varsCausingUnknownKeys []string
}

// IsIterable returns if the attribute can be ranged over.
func (attr *Attribute) IsIterable() bool {
	if attr == nil {
		return false
	}

	return attr.Value().Type().IsCollectionType() || attr.Value().Type().IsObjectType() || attr.Value().Type().IsMapType() || attr.Value().Type().IsListType() || attr.Value().Type().IsSetType() || attr.Value().Type().IsTupleType()
}

// AsInt returns the Attribute value as a int64. If the cty.Value is not a type
// that can be converted to integer, this method returns 0.
func (attr *Attribute) AsInt() int64 {
	if attr == nil {
		return 0
	}

	v := attr.Value()
	if v.IsNull() {
		return 0
	}

	if v.Type() != cty.Number {
		var err error
		v, err = convert.Convert(v, cty.Number)
		if err != nil {
			attr.Logger.Debug().Err(err).Msgf("could not return attribute value of type %s as cty.Number", v.Type())
			return 0
		}
	}

	var i int64
	err := gocty.FromCtyValue(v, &i)
	if err != nil {
		attr.Logger.Debug().Err(err).Msg("could not return attribute value as int64")
	}

	return i
}

// AsString returns the Attribute value as a string. If the cty.Value is not a type
// that can be converted to string, this method returns an empty string.
func (attr *Attribute) AsString() string {
	if attr == nil {
		return ""
	}

	v := attr.Value()
	if v.IsNull() {
		return ""
	}

	if v.Type() != cty.String {
		var err error
		v, err = convert.Convert(v, cty.String)
		if err != nil {
			attr.Logger.Debug().Err(err).Msgf("could not return attribute value of type %s as cty.String", v.Type())
			return ""
		}
	}

	var s string
	err := gocty.FromCtyValue(v, &s)
	if err != nil {
		attr.Logger.Debug().Err(err).Msg("could not return attribute value as string")
	}

	return s
}

// Value returns the Attribute with the underlying hcl.Expression of the hcl.Attribute evaluated with
// the Attribute Context. This returns a cty.Value with the values filled from any variables or references
// that the Context carries.
func (attr *Attribute) Value() cty.Value {
	if attr == nil {
		return cty.DynamicVal
	}

	attr.Logger.Trace().Msg("fetching attribute value")
	var val cty.Value
	if attr.IsGraph {
		val = attr.graphValue()
	} else {
		val = attr.value(0)
	}
	attr.previousValue = val

	return val
}

var missingVarPrefixes = []string{
	"data",
	"var",
	"module",
	"local",
}

// ReferencesCausingUnknownKeys returns a list of missing references if the attribute is an object without fully known keys.
// For example, this will return []string{"var.default_tags"} if a value is set to the result of `merge({"x": "y"}, var.default_tags)`
// where the value of `var.default_tags` is not known at evaluation time.
func (attr *Attribute) ReferencesCausingUnknownKeys() []string {
	if attr == nil {
		return nil
	}
	_ = attr.value(0)
	if len(attr.varsCausingUnknownKeys) == 0 {
		return nil
	}
	unique := make(map[string]struct{})
	for _, v := range attr.varsCausingUnknownKeys {
		var valid bool
		for _, prefix := range missingVarPrefixes {
			if strings.HasPrefix(v, prefix+".") {
				valid = true
				break
			}
		}
		if !valid {
			continue
		}
		unique[v] = struct{}{}
	}
	attr.varsCausingUnknownKeys = nil
	for k := range unique {
		attr.varsCausingUnknownKeys = append(attr.varsCausingUnknownKeys, k)
	}
	sort.Strings(attr.varsCausingUnknownKeys)
	return attr.varsCausingUnknownKeys
}

// ProvidersValue retrieves the value of the attribute with special handling need for module.providers
// blocks: Keys in the providers block are converted to literal values, then the attr.Value() is returned.
func (attr *Attribute) ProvidersValue() cty.Value {
	if origExpr, ok := attr.HCLAttr.Expr.(*hclsyntax.ObjectConsExpr); ok {
		newExpr := &hclsyntax.ObjectConsExpr{}

		for _, item := range origExpr.Items {
			if origKeyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr); ok {
				key := traversalAsString(origKeyExpr.AsTraversal())

				literalKey := &hclsyntax.LiteralValueExpr{
					Val: cty.StringVal(key),
				}

				newExpr.Items = append(newExpr.Items, hclsyntax.ObjectConsItem{
					KeyExpr:   literalKey,
					ValueExpr: item.ValueExpr,
				})
			} else {
				newExpr.Items = append(newExpr.Items, item)
			}
		}

		attr.HCLAttr.Expr = newExpr
	}

	return attr.Value()
}

// DecodeProviders decodes the providers block into a map of provider names to provider aliases.
// This is used by the graph evaluator to make sure the correct edges are created when providers are
// inherited from parent modules.
func (attr *Attribute) DecodeProviders() map[string]string {
	providers := make(map[string]string)

	if origExpr, ok := attr.HCLAttr.Expr.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range origExpr.Items {
			keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
			if !ok {
				continue
			}

			valExpr, ok := item.ValueExpr.(*hclsyntax.ScopeTraversalExpr)
			if !ok {
				continue
			}

			key := traversalAsString(keyExpr.AsTraversal())
			val := traversalAsString(valExpr.AsTraversal())

			providers[key] = val
		}
	}

	return providers
}

// HasChanged returns if the Attribute Value has changed since Value was last called.
func (attr *Attribute) HasChanged() (change bool) {
	if attr == nil {
		return false
	}

	defer func() {
		e := recover()
		if e != nil {
			attr.Logger.Debug().Msgf("HasChanged panicked with cty.Value comparison %s", e)
			change = true
		}
	}()

	previous := attr.previousValue
	var current cty.Value
	if attr.IsGraph {
		current = attr.graphValue()
	} else {
		current = attr.value(0)
	}
	return !previous.RawEquals(current)
}

// designed only to work in situations where the expression may result in an object
func extractTraversalStringsFromExpr(expr hcl.Expression) []string {
	switch v := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		return []string{traversalAsString(v.AsTraversal())}
	case *hclsyntax.IndexExpr:
		return extractTraversalStringsFromExpr(v.Collection)
	default:
		return nil
	}
}

func (attr *Attribute) value(retry int) (ctyVal cty.Value) {
	defer func() {
		if err := recover(); err != nil {
			trace := debug.Stack()
			attr.Logger.Debug().Msgf("could not evaluate value for attr: %s. This is most likely an issue in the underlying hcl/go-cty libraries and can be ignored, but we log the stacktrace for debugging purposes. Err: %s\n%s", attr.Name(), err, trace)
		}
	}()

	var diag hcl.Diagnostics
	ctyVal, diag = attr.HCLAttr.Expr.Value(attr.Ctx.Inner())
	if diag.HasErrors() {
		mockedVal := cty.StringVal(fmt.Sprintf("%s-%s", attr.Name(), mock.Identifier))
		if attr.newMock != nil {
			mockedVal = attr.newMock(attr)
		}

		if retry > 2 {
			return mockedVal
		}

		ctx := attr.Ctx.Inner()
		exp, replaced := mockFunctionCallArgs(attr.HCLAttr.Expr, diag, mockedVal)
		val, err := exp.Value(ctx)
		if !err.HasErrors() {
			attr.varsCausingUnknownKeys = append(attr.varsCausingUnknownKeys, replaced...)
			return val
		} else if len(err) < len(diag) {
			// handle cases where there is a further error inside a function param, e.g. merge({"x": var.undefined}, var.whatever)
			attr.varsCausingUnknownKeys = append(attr.varsCausingUnknownKeys, replaced...)
		}

		for _, d := range diag {

			if d.Summary == unknownVariableDiagnostic && !ctyVal.IsKnown() {
				missing := extractTraversalStringsFromExpr(d.Expression)
				for _, m := range missing {
					if !strings.HasSuffix(m, ".id") && !strings.HasSuffix(m, ".arn") {
						attr.varsCausingUnknownKeys = append(attr.varsCausingUnknownKeys, m)
					}
				}
			}

			// if the diagnostic summary indicates that we were the attribute we attempted to fetch is unsupported
			// this is likely from a Terraform attribute that is built from the provider. We then try and build
			// a mocked attribute so that the module evaluation isn't harmed.
			var shouldRetry bool
			if d.Summary == missingAttributeDiagnostic {
				badVariables := attr.findBadVariablesFromExpression(attr.HCLAttr.Expr)
				if badVariables == nil {
					badVariables = d.Expression.Variables()
				}

				shouldRetry = true
				for _, traversal := range badVariables {
					traverseVarAndSetCtx(ctx, traversal, mockedVal)
				}
			}

			if d.Summary == valueIsNonIterableDiagnostic {
				badVariables := d.Expression.Variables()
				shouldRetry = true
				for _, traversal := range badVariables {
					// let's first try and find the actual value for this bad variable.
					// If it has an actual value let's use that to pass into the list.
					val, _ := traversal.TraverseAbs(ctx)
					if val.IsNull() {
						val = mockedVal
					}

					list := cty.TupleVal([]cty.Value{val})
					traverseVarAndSetCtx(ctx, traversal, list)
				}
			}

			if d.Summary == invalidFunctionArgumentDiagnostic {
				badVariables := d.Expression.Variables()
				shouldRetry = true

				// Parse out the friendly name of the variable from the diagnostic message.
				pp := strings.Split(d.Detail, ":")
				friendlyName := strings.Trim(strings.TrimSuffix(pp[len(pp)-1], "required."), " ")

				for _, traversal := range badVariables {
					// let's first try and find the actual value for this bad variable.
					// If it has an actual value let's use that to pass into the list.
					val, _ := traversal.TraverseAbs(ctx)
					if val.IsNull() {
						val = mockedVal
					}

					if strings.HasPrefix(friendlyName, "list of") {
						val = cty.ListVal([]cty.Value{val})
					} else if strings.HasPrefix(friendlyName, "set of") {
						val = cty.SetVal([]cty.Value{val})
					} else if friendlyName == "tuple" {
						val = cty.TupleVal([]cty.Value{val})
					}

					traverseVarAndSetCtx(ctx, traversal, val)
				}
			}

			// now that we've built a mocked attribute on the global context let's try and retrieve the value once again.
			if shouldRetry {
				return attr.value(retry + 1)
			}
		}

		if attr.Verbose {
			attr.Logger.Debug().Msgf("error diagnostic return from evaluating %s err: %s", attr.HCLAttr.Name, diag.Error())
		}
	}

	return ctyVal
}

// mockFunctionCallArgs attempts to resolve remove bad expressions from hclsyntax.FunctionCallExpr args.
// This function will be called recursively, finding all functions and checking if their args match
// the bad Expressions listed in the diagnostics. This function, currently, only traverses FunctionCallExpr and ObjectCallExpr.
// More complex Expressions could be added in the future is we deem this a better way of mocking out values/expressions
// that cause evaluation to fail.
func mockFunctionCallArgs(expr hcl.Expression, diagnostics hcl.Diagnostics, mockedVal cty.Value) (hclsyntax.Expression, []string) {
	var replaced []string
	switch t := expr.(type) {
	case *hclsyntax.FunctionCallExpr:
		newArgs := make([]hclsyntax.Expression, len(t.Args))

		// loop through the function call args to get the bad expression once we've found
		// the expression which has a diagnostic let's set it as a literal value so that
		// when we evaluate the function again we don't have issues any mocked values of
		// diagnostics.
		for i, exp := range t.Args {
			var found bool
			for _, d := range diagnostics {
				if exp == d.Expression {
					// @TODO we can probably improve this by checking the function name and assigning
					// a correct mockedVal for the given function. e.g. array functions will expect a
					// list/tuple
					newArgs[i] = &hclsyntax.LiteralValueExpr{
						Val: mockedVal,
					}
					replaced = append(replaced, extractTraversalStringsFromExpr(exp)...)
					found = true
					break
				}
			}

			if !found {
				var moreReplaced []string
				newArgs[i], moreReplaced = mockFunctionCallArgs(exp, diagnostics, mockedVal)
				replaced = append(replaced, moreReplaced...)
			}
		}

		return &hclsyntax.FunctionCallExpr{
			Name:            t.Name,
			Args:            newArgs,
			ExpandFinal:     t.ExpandFinal,
			NameRange:       t.NameRange,
			OpenParenRange:  t.OpenParenRange,
			CloseParenRange: t.CloseParenRange,
		}, replaced
	case *hclsyntax.ObjectConsExpr:
		newItems := make([]hclsyntax.ObjectConsItem, len(t.Items))
		for i, item := range t.Items {
			// we don't care about replacements here, because an ObjectConsExpr already has fully known keys
			vExpr, _ := mockFunctionCallArgs(item.ValueExpr, diagnostics, mockedVal)
			newItems[i] = hclsyntax.ObjectConsItem{
				KeyExpr:   item.KeyExpr,
				ValueExpr: vExpr,
			}
		}

		return &hclsyntax.ObjectConsExpr{
			Items:     newItems,
			SrcRange:  t.SrcRange,
			OpenRange: t.OpenRange,
		}, replaced

	}

	if v, ok := expr.(hclsyntax.Expression); ok {
		return v, replaced
	}

	replaced = append(replaced, extractTraversalStringsFromExpr(expr)...)

	return &hclsyntax.LiteralValueExpr{
		Val: mockedVal,
	}, replaced
}

func (attr *Attribute) graphValue() (ctyVal cty.Value) {
	defer func() {
		if err := recover(); err != nil {
			trace := debug.Stack()
			attr.Logger.Debug().Msgf("could not evaluate value for attr: %s. This is most likely an issue in the underlying hcl/go-cty libraries and can be ignored, but we log the stacktrace for debugging purposes. Err: %s\n%s", attr.Name(), err, trace)
		}
	}()

	var diag hcl.Diagnostics
	ctyVal, diag = attr.HCLAttr.Expr.Value(attr.Ctx.Inner())
	if diag.HasErrors() {
		if !ctyVal.IsKnown() {
			for _, d := range diag {
				if d.Summary == unknownVariableDiagnostic {
					attr.varsCausingUnknownKeys = append(attr.varsCausingUnknownKeys, extractTraversalStringsFromExpr(d.Expression)...)
					break
				}
			}
		}
		mockedVal := cty.StringVal(fmt.Sprintf("%s-%s", attr.Name(), mock.Identifier))
		if attr.newMock != nil {
			mockedVal = attr.newMock(attr)
		}

		ctx := attr.Ctx.Inner()
		exp := attr.HCLAttr.Expr
		var val cty.Value
		// call the mock function in a loop to try and resolve all the bad expressions.
		// This is done because one bad expression replacement could cause another
		// expression to fail.
		for range 3 {
			exp = mockExpressionCalls(exp, diag, mockedVal)
			val, diag = exp.Value(ctx)
			if !diag.HasErrors() {
				return val
			}
		}
	}

	return ctyVal
}

// LiteralBoolValueExpression is a wrapper around any hcl.Expression that returns
// a literal bool value. This is use to evaluate mocked expressions that are used
// in conditional expressions. It turns any non bool literal value into a bool
// false value.
type LiteralBoolValueExpression struct {
	// we embed the hclsyntax.LiteralValueExpr as the hcl.Expression interface
	// has an unexported method that we need to implement.
	*hclsyntax.LiteralValueExpr

	Expression hcl.Expression
}

// Value returns the value of the expression. If the expression is not a literal
// bool value, this returns false.
func (e *LiteralBoolValueExpression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	val, diag := e.Expression.Value(ctx)
	if diag.HasErrors() {
		return cty.BoolVal(false), nil
	}

	if val.Type() != cty.Bool {
		return cty.BoolVal(false), nil
	}

	return val, nil
}

type LiteralValueCollectionExpression struct {
	// we embed the hclsyntax.LiteralValueExpr as the hcl.Expression interface
	// has an unexported method that we need to implement.
	*hclsyntax.LiteralValueExpr
	Expression  hcl.Expression
	MockedValue cty.Value
}

func newLiteralValueCollectionExpression(mockedVal cty.Value, expr hclsyntax.Expression) *LiteralValueCollectionExpression {
	return &LiteralValueCollectionExpression{
		LiteralValueExpr: &hclsyntax.LiteralValueExpr{Val: cty.ListVal([]cty.Value{mockedVal})},
		Expression:       expr,
		MockedValue:      mockedVal,
	}
}

func (e *LiteralValueCollectionExpression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	val, diag := e.Expression.Value(ctx)
	if diag.HasErrors() {
		return cty.ListValEmpty(cty.String), nil
	}

	if !val.CanIterateElements() {
		return cty.ListValEmpty(cty.String), nil
	}

	return val, nil
}

type LiteralValueIndexExpression struct {
	// we embed the hclsyntax.LiteralValueExpr as the hcl.Expression interface
	// has an unexported method that we need to implement.
	*hclsyntax.LiteralValueExpr
	Expression  *hclsyntax.IndexExpr
	MockedValue cty.Value
}

func newLiteralValueIndexExpression(mockedVal cty.Value, expr *hclsyntax.IndexExpr) *LiteralValueIndexExpression {
	return &LiteralValueIndexExpression{
		LiteralValueExpr: &hclsyntax.LiteralValueExpr{Val: cty.ListValEmpty(cty.String)},
		Expression:       expr,
		MockedValue:      mockedVal,
	}
}

func (e *LiteralValueIndexExpression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	val, diag := e.Expression.Value(ctx)
	for _, d := range diag {
		// if the diagnostic is an invalid index, we should try and get the first element
		// of the collection since it should at least have the same expected type
		if d.Summary == "Invalid index" {
			col, colDiag := e.Expression.Collection.Value(ctx)

			if !colDiag.HasErrors() && col.CanIterateElements() {
				it := col.ElementIterator()
				if it.Next() {
					_, v := it.Element()
					return v, nil
				}
			}
		}
	}

	// For other diagnostics we just return the mocked value
	// as we can't determine the correct value.
	if diag.HasErrors() {
		return e.MockedValue, nil
	}

	return val, nil
}

// mockExpressionCalls attempts to resolve remove bad expressions for the given diagnostics.
// This function will be called recursively, finding all Expressions and checking if
// the expression matches the diagnostic. If the expression matches the diagnostic, we replace the expression with a mocked value.
func mockExpressionCalls(expr hcl.Expression, diagnostics hcl.Diagnostics, mockedVal cty.Value) hclsyntax.Expression {
	switch t := expr.(type) {
	case nil:
		return nil
	case *hclsyntax.FunctionCallExpr:
		// if the diagnostic is with the function call let's replace it with a completely mocked value.
		for _, d := range diagnostics {
			if t == d.Expression {
				return &hclsyntax.LiteralValueExpr{
					Val: mockedVal,
				}
			}
		}

		newArgs := make([]hclsyntax.Expression, len(t.Args))
		for i, exp := range t.Args {
			newArgs[i] = mockExpressionCalls(exp, diagnostics, mockedVal)
		}

		return &hclsyntax.FunctionCallExpr{
			Name:            t.Name,
			Args:            newArgs,
			ExpandFinal:     t.ExpandFinal,
			NameRange:       t.NameRange,
			OpenParenRange:  t.OpenParenRange,
			CloseParenRange: t.CloseParenRange,
		}
	case *hclsyntax.ObjectConsExpr:
		newItems := make([]hclsyntax.ObjectConsItem, len(t.Items))
		for i, item := range t.Items {
			newItems[i] = hclsyntax.ObjectConsItem{
				KeyExpr:   mockExpressionCalls(item.KeyExpr, diagnostics, mockedVal),
				ValueExpr: mockExpressionCalls(item.ValueExpr, diagnostics, mockedVal),
			}
		}

		return &hclsyntax.ObjectConsExpr{
			Items:     newItems,
			SrcRange:  t.SrcRange,
			OpenRange: t.OpenRange,
		}
	case *hclsyntax.ConditionalExpr:
		return &hclsyntax.ConditionalExpr{
			Condition:   mockBoolExpressionCall(t.Condition, diagnostics, mockedVal),
			TrueResult:  mockExpressionCalls(t.TrueResult, diagnostics, mockedVal),
			FalseResult: mockExpressionCalls(t.FalseResult, diagnostics, mockedVal),
			SrcRange:    t.SrcRange,
		}
	case *hclsyntax.TemplateWrapExpr:
		return &hclsyntax.TemplateWrapExpr{
			Wrapped:  mockExpressionCalls(t.Wrapped, diagnostics, mockedVal),
			SrcRange: t.SrcRange,
		}
	case *hclsyntax.TemplateExpr:
		for _, d := range diagnostics {
			if t == d.Expression {
				return &hclsyntax.LiteralValueExpr{
					Val: mockedVal,
				}
			}
		}

		newParts := make([]hclsyntax.Expression, len(t.Parts))
		for i, part := range t.Parts {
			newParts[i] = mockExpressionCalls(part, diagnostics, mockedVal)
		}

		return &hclsyntax.TemplateExpr{
			Parts:    newParts,
			SrcRange: t.SrcRange,
		}
	case *hclsyntax.TupleConsExpr:
		newExprs := make([]hclsyntax.Expression, len(t.Exprs))
		for i, exp := range t.Exprs {
			newExprs[i] = mockExpressionCalls(exp, diagnostics, mockedVal)
		}

		return &hclsyntax.TupleConsExpr{
			Exprs:     newExprs,
			SrcRange:  t.SrcRange,
			OpenRange: t.OpenRange,
		}
	case *hclsyntax.IndexExpr:
		expr := &hclsyntax.IndexExpr{
			Collection:   newLiteralValueCollectionExpression(mockedVal, mockExpressionCalls(t.Collection, diagnostics, mockedVal)),
			Key:          mockExpressionCalls(t.Key, diagnostics, mockedVal),
			SrcRange:     t.SrcRange,
			OpenRange:    t.OpenRange,
			BracketRange: t.BracketRange,
		}
		return newLiteralValueIndexExpression(mockedVal, expr)
	case *hclsyntax.ForExpr:
		return &hclsyntax.ForExpr{
			KeyVar:     t.KeyVar,
			ValVar:     t.ValVar,
			CollExpr:   newLiteralValueCollectionExpression(mockedVal, mockExpressionCalls(t.CollExpr, diagnostics, mockedVal)),
			KeyExpr:    mockExpressionCalls(t.KeyExpr, diagnostics, mockedVal),
			ValExpr:    mockExpressionCalls(t.ValExpr, diagnostics, mockedVal),
			CondExpr:   mockBoolExpressionCall(t.CondExpr, diagnostics, mockedVal),
			Group:      t.Group,
			SrcRange:   t.SrcRange,
			OpenRange:  t.OpenRange,
			CloseRange: t.CloseRange,
		}
	case *hclsyntax.ObjectConsKeyExpr:
		return &hclsyntax.ObjectConsKeyExpr{
			Wrapped:         mockExpressionCalls(t.Wrapped, diagnostics, mockedVal),
			ForceNonLiteral: t.ForceNonLiteral,
		}
	case *hclsyntax.SplatExpr:
		return &hclsyntax.SplatExpr{
			Source:      mockExpressionCalls(t.Source, diagnostics, mockedVal),
			Each:        mockExpressionCalls(t.Each, diagnostics, mockedVal),
			Item:        t.Item,
			SrcRange:    t.SrcRange,
			MarkerRange: t.MarkerRange,
		}
	case *hclsyntax.BinaryOpExpr:
		// Logical operators (|| and &&) expect bools on both sides, so we need to mock the expression
		// with a bool value.
		impl := t.Op.Impl
		params := impl.Params()

		var lhsVal hclsyntax.Expression
		var rhsVal hclsyntax.Expression

		if len(params) > 0 && params[0].Type == cty.Bool {
			lhsVal = mockBoolExpressionCall(t.LHS, diagnostics, mockedVal)
		} else {
			lhsVal = mockExpressionCalls(t.LHS, diagnostics, mockedVal)
		}

		if len(params) > 1 && params[1].Type == cty.Bool {
			rhsVal = mockBoolExpressionCall(t.RHS, diagnostics, mockedVal)
		} else {
			rhsVal = mockExpressionCalls(t.RHS, diagnostics, mockedVal)
		}

		return &hclsyntax.BinaryOpExpr{
			LHS:      lhsVal,
			Op:       t.Op,
			RHS:      rhsVal,
			SrcRange: t.SrcRange,
		}
	case *hclsyntax.UnaryOpExpr:
		// The ! operator expects a bool on the right side, so we need to mock the expression
		// with a bool value.
		impl := t.Op.Impl
		params := impl.Params()

		var val hclsyntax.Expression
		if len(params) > 0 && params[0].Type == cty.Bool {
			val = mockBoolExpressionCall(t.Val, diagnostics, mockedVal)
		} else {
			val = mockExpressionCalls(t.Val, diagnostics, mockedVal)
		}

		return &hclsyntax.UnaryOpExpr{
			Op:          t.Op,
			Val:         val,
			SrcRange:    t.SrcRange,
			SymbolRange: t.SymbolRange,
		}
	case *hclsyntax.TemplateJoinExpr:
		return &hclsyntax.TemplateJoinExpr{
			Tuple: mockExpressionCalls(t.Tuple, diagnostics, mockedVal),
		}
	case *hclsyntax.ParenthesesExpr:
		return &hclsyntax.ParenthesesExpr{
			Expression: mockExpressionCalls(t.Expression, diagnostics, mockedVal),
			SrcRange:   t.SrcRange,
		}
	case *hclsyntax.RelativeTraversalExpr:
		for _, d := range diagnostics {
			if t == d.Expression {
				return &hclsyntax.LiteralValueExpr{
					Val: mockedVal,
				}
			}
		}

		return &hclsyntax.RelativeTraversalExpr{
			Source:    mockExpressionCalls(t.Source, diagnostics, mockedVal),
			Traversal: t.Traversal,
			SrcRange:  t.SrcRange,
		}
	case *hclsyntax.AnonSymbolExpr:
	case *hclsyntax.LiteralValueExpr:
	case *hclsyntax.ScopeTraversalExpr:
		for _, d := range diagnostics {
			if t == d.Expression {
				if d.Summary == "Iteration over non-iterable value" {
					return &hclsyntax.LiteralValueExpr{
						Val: cty.ListValEmpty(cty.String),
					}
				}

				return &hclsyntax.LiteralValueExpr{
					Val: mockedVal,
				}
			}
		}

		return t
	}

	if v, ok := expr.(hclsyntax.Expression); ok {
		return v
	}

	return &hclsyntax.LiteralValueExpr{
		Val: mockedVal,
	}
}

func mockBoolExpressionCall(expression hclsyntax.Expression, diagnostics hcl.Diagnostics, mockedVal cty.Value) hclsyntax.Expression {
	if expression == nil {
		return nil
	}

	var condition hclsyntax.Expression
	for _, d := range diagnostics {
		if expression == d.Expression {
			condition = &hclsyntax.LiteralValueExpr{
				Val: cty.BoolVal(false),
			}
			break
		}
	}

	if condition == nil {
		condition = mockExpressionCalls(expression, diagnostics, mockedVal)
	}

	return &LiteralBoolValueExpression{Expression: condition, LiteralValueExpr: &hclsyntax.LiteralValueExpr{Val: cty.BoolVal(false)}}
}

// traverseVarAndSetCtx uses the hcl traversal to build a mocked attribute on the evaluation context.
// hcl Traversals from missing are normally provided in the following manner:
//
//  1. The root traversal or TraverseRoot fetches the top level reference for the block. We use this traversal to
//     determine which ctx we use. We loop through the list of EvaluationContext until we find an entry matching the
//     reference. If there is none, we exit, this shouldn't happen and is likely an indicator of a bug.
//  2. The remaining attribute traversals or TraverseAttr. These use the value fetched from the context by the TraverseRoot
//     to find the value of the attribute the expression is trying to evaluate. In our case this is the attribute that
//     we need to populate with a mocked value.
//
// Once we've found the missing attribute we set a mocked value and return. This value should now be available for
// the entire context evaluation as ctx is share across all blocks.
func traverseVarAndSetCtx(ctx *hcl.EvalContext, traversal hcl.Traversal, mock cty.Value) {
	var rootName string
	for _, traverser := range traversal {
		if r, ok := traverser.(hcl.TraverseRoot); ok {
			rootName = r.Name
			break
		}
	}

	if rootName == "" {
		return
	}

	ctx = findCorrectCtx(ctx, rootName)
	if ctx == nil {
		return
	}

	ob := ctx.Variables[rootName]
	if ob.IsNull() || !ob.IsKnown() {
		ob = cty.ObjectVal(make(map[string]cty.Value))
	}

	ctx.Variables[rootName] = buildObject(traversal, ob, mock, 0)
}

// buildObject builds an attribute map from the traversal. It fills any missing attributes that are
// defined by the traversal.
func buildObject(traversal hcl.Traversal, value cty.Value, mock cty.Value, i int) cty.Value {
	if i > len(traversal)-1 {
		return value
	}

	traverser := traversal[i]
	valueMap := value.AsValueMap()
	if valueMap == nil {
		valueMap = make(map[string]cty.Value)
	}

	// traverse splat is a special holding type which means we want to traverse all the attributes on the map.
	if _, ok := traverser.(hcl.TraverseSplat); ok {
		for k, v := range valueMap {
			if v.Type().IsObjectType() {
				valueMap[k] = buildObject(traversal, v, mock, i+1)
				continue
			}

			valueMap[k] = v
		}

		return cty.ObjectVal(valueMap)
	}

	if index, ok := traverser.(hcl.TraverseIndex); ok {
		kc, err := convert.Convert(index.Key, cty.String)
		if err != nil {
			kc = cty.StringVal("0")
		}

		k := kc.AsString()

		if vv, exists := valueMap[k]; exists {
			valueMap[k] = buildObject(traversal, vv, mock, i+1)
			return cty.ObjectVal(valueMap)
		}

		if len(traversal)-1 == i {
			valueMap[k] = mock
		} else {
			valueMap[k] = buildObject(traversal, cty.ObjectVal(make(map[string]cty.Value)), mock, i+1)
		}

		return cty.ObjectVal(valueMap)
	}

	if v, ok := traverser.(hcl.TraverseAttr); ok {
		if len(traversal)-1 == i {
			// if the attribute already exists, and we're not setting a list value
			// then we should return here. It's most likely that we weren't able to
			// get the full variable calls for the context, so resetting the value could
			// be harmful.
			if _, exists := valueMap[v.Name]; exists && mock.Type() == cty.String {
				return value
			}

			valueMap[v.Name] = mock
			return cty.ObjectVal(valueMap)
		}

		if vv, exists := valueMap[v.Name]; exists {
			if isList(vv) {
				items := make([]cty.Value, vv.LengthInt())
				it := vv.ElementIterator()
				for it.Next() {
					key, sourceItem := it.Element()
					val := buildObject(traversal, sourceItem, mock, i+1)
					i, _ := key.AsBigFloat().Int64()
					items[i] = val
				}
				valueMap[v.Name] = cty.TupleVal(items)
				return cty.ObjectVal(valueMap)
			}

			next := traversal[i+1]
			if _, ok := next.(hcl.TraverseIndex); ok {
				if !isList(vv) {
					vv = cty.TupleVal([]cty.Value{vv})
				}
			}

			valueMap[v.Name] = buildObject(traversal, vv, mock, i+1)
			return cty.ObjectVal(valueMap)
		}

		valueMap[v.Name] = buildObject(traversal, cty.ObjectVal(make(map[string]cty.Value)), mock, i+1)
		return cty.ObjectVal(valueMap)
	}

	return buildObject(traversal, value, mock, i+1)
}

// findCorrectCtx uses name to find the correct context to target. findCorrectCtx returns the first
// context that contains an instance of name. If no entries of name are found findCorrectCtx returns nil.
func findCorrectCtx(ctx *hcl.EvalContext, name string) *hcl.EvalContext {
	thisCtx := ctx
	for thisCtx != nil {
		if thisCtx.Variables == nil {
			thisCtx = thisCtx.Parent()
			continue
		}
		_, exists := thisCtx.Variables[name]
		if exists {
			return thisCtx
		}

		thisCtx = thisCtx.Parent()
	}

	return thisCtx
}

// Name is a helper method to return the underlying hcl.Attribute Name
func (attr *Attribute) Name() string {
	return attr.HCLAttr.Name
}

// Equals checks that val matches the underlying Attribute cty.Type.
func (attr *Attribute) Equals(val any) bool {
	if attr.Value().Type() == cty.String {
		result := strings.EqualFold(attr.Value().AsString(), fmt.Sprintf("%v", val))
		return result
	}

	if attr.Value().Type() == cty.Bool {
		return attr.Value().True() == val
	}

	if attr.Value().Type() == cty.Number {
		checkNumber, err := gocty.ToCtyValue(val, cty.Number)
		if err != nil {
			attr.Logger.Debug().Msgf("Error converting number for equality check. %s", err)
			return false
		}
		return attr.Value().RawEquals(checkNumber)
	}

	return false
}

func (attr *Attribute) createDotReferenceFromTraversal(traversals ...hcl.Traversal) (*Reference, error) {
	var refParts []string

	for _, x := range traversals {
		for _, p := range x {
			switch part := p.(type) {
			case hcl.TraverseRoot:
				refParts = append(refParts, part.Name)
			case hcl.TraverseAttr:
				refParts = append(refParts, part.Name)
			case hcl.TraverseIndex:
				refParts[len(refParts)-1] = fmt.Sprintf("%s[%s]", refParts[len(refParts)-1], attr.getIndexValue(part))
			}
		}
	}
	return newReference(refParts)
}

func (attr *Attribute) getIndexValue(part hcl.TraverseIndex) string {
	switch part.Key.Type() {
	case cty.String:
		return fmt.Sprintf("%q", part.Key.AsString())
	case cty.Number:
		var intVal int
		if err := gocty.FromCtyValue(part.Key, &intVal); err != nil {
			attr.Logger.Debug().Err(err).Msg("could not unpack int from block index attr, returning 0")
			return "0"
		}

		return fmt.Sprintf("%d", intVal)
	default:
		attr.Logger.Debug().Msgf("could not get index value for unsupported cty type %s, returning 0", part.Key.Type())
		return "0"
	}
}

// Reference returns the pointer to a Reference struct that holds information about the Attributes
// referenced block. Reference achieves this by traversing the Attribute Expression in order to find the
// parent block. E.g. with the following HCL
//
//	resource "aws_launch_template" "foo2" {
//		name = "foo2"
//	}
//
//	resource "some_resource" "example_with_launch_template_3" {
//		...
//		name    = aws_launch_template.foo2.name
//	}
//
// The Attribute some_resource.name would have a reference of
//
//	Reference {
//		blockType: Type{
//			name:                  "resource",
//			removeTypeInReference: true,
//		}
//		typeLabel: "aws_launch_template"
//		nameLabel: "foo2"
//	}
//
// Reference is used to build up a Terraform JSON configuration file that holds information about the expressions
// and their parents. Infracost uses these references in resource evaluation to lookup connecting resource information.
func (attr *Attribute) Reference() (*Reference, error) {
	if attr == nil {
		return nil, fmt.Errorf("attribute is nil")
	}

	refs := attr.AllReferences()
	if len(refs) == 0 {
		return nil, fmt.Errorf("no references for attribute")
	}

	return refs[0], nil
}

// AllReferences returns a list of References for the given Attribute. This can include the
// main Value Reference (see Reference method) and also a list of references used in conditional
// evaluation and templating.
func (attr *Attribute) AllReferences() []*Reference {
	if attr == nil {
		return nil
	}

	return attr.referencesFromExpression(attr.HCLAttr.Expr)
}

// VerticesReferenced traverses all the Expressions used by the attribute to build a
// list of all the Blocks referenced by the Attribute.
func (attr *Attribute) VerticesReferenced(b *Block) []VertexReference {
	allRefs := attr.AllReferences()
	refs := make([]VertexReference, 0, len(allRefs))

	for _, ref := range allRefs {
		key := ref.String()

		if shouldSkipRef(b, attr, key) {
			continue
		}

		isProviderReference := usesProviderConfiguration(b) && attr.Name() == "provider"
		if isProviderReference {
			key = fmt.Sprintf("provider.%s", strings.TrimSuffix(key, "."))
		}

		modAddr := b.ModuleAddress()
		modPart, otherPart := splitModuleAddr(key)

		if modPart != "" {
			if modAddr == "" {
				modAddr = modPart
			} else {
				modAddr = fmt.Sprintf("%s.%s", modAddr, modPart)
			}
		}

		refs = append(refs, VertexReference{
			ModuleAddress: modAddr,
			AttributeName: attr.Name(),
			Key:           otherPart,
		})
	}

	return refs
}

func (attr *Attribute) referencesFromExpression(expression hcl.Expression) []*Reference {
	if attr == nil {
		return nil
	}

	var refs []*Reference
	switch t := expression.(type) {
	case *hclsyntax.FunctionCallExpr:
		for _, arg := range t.Args {
			refs = append(refs, attr.referencesFromExpression(arg)...)
		}
	case *hclsyntax.ConditionalExpr:
		refs = append(refs, attr.referencesFromExpression(t.TrueResult)...)
		refs = append(refs, attr.referencesFromExpression(t.FalseResult)...)
		refs = append(refs, attr.referencesFromExpression(t.Condition)...)
	case *hclsyntax.ScopeTraversalExpr:
		if ref, err := attr.createDotReferenceFromTraversal(t.Variables()...); err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.TemplateWrapExpr:
		refs = attr.referencesFromExpression(t.Wrapped)
	case *hclsyntax.TemplateExpr:
		for _, part := range t.Parts {
			refs = append(refs, attr.referencesFromExpression(part)...)
		}
	case *hclsyntax.TupleConsExpr:
		for _, item := range t.Exprs {
			refs = append(refs, attr.referencesFromExpression(item)...)
		}
	case *hclsyntax.RelativeTraversalExpr:
		refs = append(refs, attr.referencesFromExpression(t.Source)...)
	case *hclsyntax.IndexExpr:
		refs = append(refs, attr.referencesFromExpression(t.Collection)...)
		refs = append(refs, attr.referencesFromExpression(t.Key)...)
	case *hclsyntax.ForExpr:
		refs := attr.referencesFromExpression(t.CollExpr)
		refs = append(refs, attr.referencesFromExpression(t.KeyExpr)...)
		refs = append(refs, attr.referencesFromExpression(t.ValExpr)...)

		if t.CondExpr != nil {
			refs = append(refs, attr.referencesFromExpression(t.CondExpr)...)
		}
		return refs
	case *hclsyntax.ObjectConsExpr:
		for _, item := range t.Items {
			refs = append(refs, attr.referencesFromExpression(item.KeyExpr)...)
			refs = append(refs, attr.referencesFromExpression(item.ValueExpr)...)
		}
		return refs
	case *hclsyntax.ObjectConsKeyExpr:
		// If the traversal is of length one it is treated as a string by Terraform.
		// Otherwise it could be a reference. For example:
		//
		// providers = {
		//   aws = aws.alias
		// }
		//
		// In this case the traversal of the key expression would be of length 1 and
		// we would treat it as a string.
		//
		// TODO: Although this helps, I think we still need some way of totally ignoring keys for
		// the providers attribute of module calls since they can contain a '.' and therefore have
		// a traversal, but are a special case that should be treated as strings.
		wrapped, ok := t.Wrapped.(*hclsyntax.ScopeTraversalExpr)
		if ok && len(wrapped.Traversal) > 1 {
			refs = append(refs, attr.referencesFromExpression(t.Wrapped)...)
		}

		return refs
	case *hclsyntax.SplatExpr:
		refs = append(refs, attr.referencesFromExpression(t.Source)...)
		return refs
	case *hclsyntax.BinaryOpExpr:
		refs = append(refs, attr.referencesFromExpression(t.LHS)...)
		refs = append(refs, attr.referencesFromExpression(t.RHS)...)
		return refs
	case *hclsyntax.UnaryOpExpr:
		refs = append(refs, attr.referencesFromExpression(t.Val)...)
		return refs
	case *hclsyntax.TemplateJoinExpr:
		refs = append(refs, attr.referencesFromExpression(t.Tuple)...)
	case *hclsyntax.ParenthesesExpr:
		refs = append(refs, attr.referencesFromExpression(t.Expression)...)
	case *hclsyntax.AnonSymbolExpr:
		ref, err := attr.createDotReferenceFromTraversal(t.Variables()...)
		if err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.LiteralValueExpr:
		attr.Logger.Trace().Msgf("cannot create references from %T as it is a literal value and will not contain refs", t)
	default:
		name := fmt.Sprintf("%T", t)
		if strings.HasPrefix(name, "*hclsyntax") {
			// if we get here then that means we have encountered an expression type that we don't support.
			// Adding support for newly added expressions is critical to graph evaluation, so let's log an error.
			attr.Logger.Error().Msgf("cannot create references for unsupported expression type %q", name)
		} else {
			attr.Logger.Debug().Msgf("cannot create references for expression type: %q", name)
		}
	}

	return refs
}

// findBadVariablesFromExpression attempts to find the variables that are missing by calling the underlying expressions
// and checking if they have any missing attributes diagnostics. findBadVariablesFromExpression is a fallback method
// as normally Diagnostics return the variables we need. However, in cases where the expressions are complex (e.g.
// a splat expression within a function call) the Diagnostics will only have variable information from the last expression.
// Meaning that in many cases they won't actually contain the problem variables and calling diag.Variables() will return nil.
func (attr *Attribute) findBadVariablesFromExpression(expression hcl.Expression) []hcl.Traversal {
	var badVars []hcl.Traversal
	ctx := attr.Ctx.Inner()
	switch t := expression.(type) {
	case *hclsyntax.ForExpr:
		// if there are bad vars in the collection we need to evaluate these first
		badVars = append(badVars, attr.findBadVariablesFromExpression(t.CollExpr)...)
		if badVars != nil {
			return badVars
		}

		collVal, _ := t.CollExpr.Value(ctx)
		if !isList(collVal) {
			collVal = cty.TupleVal([]cty.Value{collVal})
		}
		it := collVal.ElementIterator()

		for it.Next() {
			k, v := it.Element()
			childCtx := ctx.NewChild()
			childCtx.Variables = map[string]cty.Value{}
			if t.KeyVar != "" {
				childCtx.Variables[t.KeyVar] = k
			}
			childCtx.Variables[t.ValVar] = v

			if t.CondExpr != nil {
				_, diags := t.CondExpr.Value(childCtx)
				if isAttrMissing(diags) {
					trav := findBadTraversal(childCtx, t.CondExpr)

					if trav.RootName() == t.ValVar {
						abs, _ := hcl.AbsTraversalForExpr(t.CollExpr)
						rels := toRelativeTraversal(trav)
						traversal := hcl.TraversalJoin(abs, rels)
						badVars = append(badVars, traversal)

						return badVars
					}
				}

			}
		}

		if t.CondExpr != nil {
			badVars = append(badVars, attr.findBadVariablesFromExpression(t.CondExpr)...)
		}

		badVars = append(badVars, attr.findBadVariablesFromExpression(t.ValExpr)...)
		return badVars
	case *hclsyntax.FunctionCallExpr:
		for _, arg := range t.Args {
			badVars = append(badVars, attr.findBadVariablesFromExpression(arg)...)
		}

		return badVars
	case *hclsyntax.ConditionalExpr:
		badVars = append(badVars, attr.findBadVariablesFromExpression(t.TrueResult)...)
		badVars = append(badVars, attr.findBadVariablesFromExpression(t.FalseResult)...)
		badVars = append(badVars, attr.findBadVariablesFromExpression(t.Condition)...)
		return badVars
	case *hclsyntax.TemplateWrapExpr:
		return attr.findBadVariablesFromExpression(t.Wrapped)
	case *hclsyntax.TemplateExpr:
		for _, part := range t.Parts {
			badVars = append(badVars, attr.findBadVariablesFromExpression(part)...)
		}

		return badVars
	case *hclsyntax.RelativeTraversalExpr:
		switch s := t.Source.(type) {
		case *hclsyntax.IndexExpr:
			ctx := attr.Ctx.Inner()
			val, diags := s.Collection.Value(ctx)
			if isAttrMissing(diags) {
				return attr.findBadVariables(s.Collection.Variables())
			}

			it := val.ElementIterator()
			for it.Next() {
				_, v := it.Element()

				_, d := t.Traversal.TraverseRel(v)
				if isAttrMissing(d) {
					traversal := s.Collection.Variables()[0]
					traversal = append(traversal, hcl.TraverseSplat{})
					traversal = append(traversal, t.Traversal...)
					badVars = append(badVars, traversal)
					return badVars
				} else {
					break
				}
			}

			return attr.findBadVariables(s.Collection.Variables())
		default:
			return attr.findBadVariablesFromExpression(t.Source)
		}
	case *hclsyntax.IndexExpr:
		return attr.findBadVariables(t.Collection.Variables())
	case *hclsyntax.SplatExpr:
		_, diag := t.Value(attr.Ctx.Inner())
		if isAttrMissing(diag) {
			baseVars := t.Variables()

			if rt, ok := t.Each.(*hclsyntax.RelativeTraversalExpr); ok {
				for i, baseVar := range baseVars {
					baseVars[i] = append(baseVar, rt.Traversal...)
				}

				badVars = append(badVars, baseVars...)
				return badVars
			}
		}
	case *hclsyntax.ObjectConsExpr:
		for _, item := range t.Items {
			badVars = append(badVars, attr.findBadVariablesFromExpression(item.KeyExpr)...)
			badVars = append(badVars, attr.findBadVariablesFromExpression(item.ValueExpr)...)
		}

		return badVars
	}

	return attr.findBadVariables(expression.Variables())
}

func findBadTraversal(ctx *hcl.EvalContext, expression hcl.Expression) hcl.Traversal {
	switch t := expression.(type) {
	case *hclsyntax.ConditionalExpr:
		if b := findBadTraversal(ctx, t.TrueResult); b != nil {
			return b
		}
		if b := findBadTraversal(ctx, t.FalseResult); b != nil {
			return b
		}
		if b := findBadTraversal(ctx, t.Condition); b != nil {
			return b
		}
	case *hclsyntax.BinaryOpExpr:
		if b := findBadTraversal(ctx, t.LHS); b != nil {
			return b
		}
		if b := findBadTraversal(ctx, t.RHS); b != nil {
			return b
		}
	}

	_, diag := expression.Value(ctx)
	if isAttrMissing(diag) {
		trav, _ := hcl.AbsTraversalForExpr(expression)
		return trav
	}

	return nil
}

func (attr *Attribute) findBadVariables(traversals []hcl.Traversal) []hcl.Traversal {
	var badVars []hcl.Traversal

	for _, traversal := range traversals {
		if traversal.IsRelative() {
			continue
		}

		_, diag := traversal.TraverseAbs(attr.Ctx.Inner())
		if isAttrMissing(diag) {
			badVars = append(badVars, traversal)
		}
	}

	return badVars
}

func isAttrMissing(diag hcl.Diagnostics) bool {
	for _, d := range diag {
		if d.Summary == missingAttributeDiagnostic {
			return true
		}
	}

	return false
}

func isList(v cty.Value) bool {
	sourceTy := v.Type()

	return sourceTy.IsTupleType() || sourceTy.IsListType() || sourceTy.IsSetType()
}

func toRelativeTraversal(traversal hcl.Traversal) hcl.Traversal {
	var ret hcl.Traversal
	for _, traverser := range traversal {
		if _, ok := traverser.(hcl.TraverseRoot); ok {
			continue
		}

		ret = append(ret, traverser)
	}

	return ret
}

func traversalAsString(traversal hcl.Traversal) string {
	buf := bytes.Buffer{}
	for _, tr := range traversal {
		switch step := tr.(type) {
		case hcl.TraverseRoot:
			buf.WriteString(step.Name)
		case hcl.TraverseAttr:
			buf.WriteString(".")
			buf.WriteString(step.Name)
		}
	}
	return buf.String()
}

func shouldSkipRef(block *Block, attr *Attribute, key string) bool {
	if key == "count.index" || key == "each.key" || key == "each.value" {
		return true
	}

	// Provider references can come through as `aws.`
	isProviderReference := usesProviderConfiguration(block) && attr.Name() == "provider"
	if !isProviderReference && strings.HasSuffix(key, ".") {
		return true
	}

	if block.parent != nil && block.parent.Type() == "variable" && block.Type() == "validation" {
		return true
	}

	return false
}

func splitModuleAddr(address string) (string, string) {
	matches := addrSplitModuleRegex.FindStringSubmatch(address)
	if len(matches) == 3 {
		return matches[1], matches[2]
	}
	return "", address
}
