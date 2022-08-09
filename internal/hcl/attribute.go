package hcl

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

var (
	missingAttributeDiagnostic   = "Unsupported attribute"
	valueIsNonIterableDiagnostic = "Iteration over non-iterable value"
)

// Attribute provides a wrapper struct around hcl.Attribute it provides
// helper methods and functionality for common interactions with hcl.Attribute.
//
// Attributes are key/value pairs that are part of a Block. For example take the following Block:
//
//		resource "aws_instance" "t3_standard" {
//		  	ami           = "fake_ami"
//  		instance_type = "t3.medium"
//
//  		credit_specification {
//    			cpu_credits = "standard"
//  		}
//		}
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
	Logger  *logrus.Entry
	// newMock generates a mock value for the attribute if it's value is missing.
	newMock func(attr *Attribute) cty.Value
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
			attr.Logger.WithError(err).Debugf("could not return attribute value of type %s as cty.Number", v.Type())
			return 0
		}
	}

	var i int64
	err := gocty.FromCtyValue(v, &i)
	if err != nil {
		attr.Logger.WithError(err).Debug("could not return attribute value as int64")
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
			attr.Logger.WithError(err).Debugf("could not return attribute value of type %s as cty.String", v.Type())
			return ""
		}
	}

	var s string
	err := gocty.FromCtyValue(v, &s)
	if err != nil {
		attr.Logger.WithError(err).Debug("could not return attribute value as string")
	}

	return s
}

// Value returns the Attribute with the underlying hcl.Expression of the hcl.Attribute evaluated with
// the Attribute Context. This returns a cty.Value with the values filled from any variables or references
// that the Context carries.
func (attr *Attribute) Value() cty.Value {
	if attr == nil {
		return cty.NilVal
	}

	attr.Logger.Debug("fetching attribute value")
	return attr.value(0)
}

func (attr *Attribute) value(retry int) (ctyVal cty.Value) {
	defer func() {
		if err := recover(); err != nil {
			attr.Logger.Debugf("could not evaluate value for attr: %s. This is most likely an issue in the underlying hcl/go-cty libraries and can be ignored, but we log the stacktrace for debugging purposes. Err: %s\n%s", attr.Name(), err, debug.Stack())
			ctyVal = cty.NilVal
		}
	}()

	var diag hcl.Diagnostics
	ctyVal, diag = attr.HCLAttr.Expr.Value(attr.Ctx.Inner())
	if diag.HasErrors() {
		mockedVal := cty.StringVal(fmt.Sprintf("%s-mock", attr.Name()))
		if attr.newMock != nil {
			mockedVal = attr.newMock(attr)
		}

		if retry > 2 {
			return mockedVal
		}

		ctx := attr.Ctx.Inner()
		for _, d := range diag {
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
					if val == cty.NilVal {
						val = mockedVal
					}

					list := cty.TupleVal([]cty.Value{val})
					traverseVarAndSetCtx(ctx, traversal, list)
				}
			}

			// now that we've built a mocked attribute on the global context let's try and retrieve the value once again.
			if shouldRetry {
				return attr.value(retry + 1)
			}
		}

		if attr.Verbose {
			attr.Logger.Debugf("error diagnostic return from evaluating %s err: %s", attr.HCLAttr.Name, diag.Error())
		}
	}

	return ctyVal
}

// traverseVarAndSetCtx uses the hcl traversal to build a mocked attribute on the evaluation context.
// hcl Traversals from missing are normally provided in the following manner:
//
// 1. The root traversal or TraverseRoot fetches the top level reference for the block. We use this traversal to
//    determine which ctx we use. We loop through the list of EvaluationContext until we find an entry matching the
//    reference. If there is none, we exit, this shouldn't happen and is likely an indicator of a bug.
// 2. The remaining attribute traversals or TraverseAttr. These use the value fetched from the context by the TraverseRoot
//    to find the value of the attribute the expression is trying to evaluate. In our case this is the attribute that
//    we need to populate with a mocked value.
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

	ob := ctx.Variables[rootName].AsValueMap()
	if ob == nil {
		ob = make(map[string]cty.Value)
	}

	ob = buildObject(traversal, ob, mock, 0)
	ctx.Variables[rootName] = cty.ObjectVal(ob)
}

// buildObject builds an attribute map from the traversal. It fills any missing attributes that are
// defined by the traversal.
func buildObject(traversal hcl.Traversal, ob map[string]cty.Value, mock cty.Value, i int) map[string]cty.Value {
	if i > len(traversal)-1 {
		return ob
	}

	traverser := traversal[i]

	// traverse splat is a special holding type which means we want to traverse all the attributes on the map.
	if _, ok := traverser.(hcl.TraverseSplat); ok {
		for k, v := range ob {
			if v.Type().IsObjectType() {
				valueMap := v.AsValueMap()
				ob[k] = cty.ObjectVal(buildObject(traversal, valueMap, mock, i+1))
				continue
			}

			ob[k] = v
		}

		return ob
	}

	if index, ok := traverser.(hcl.TraverseIndex); ok {
		kc, err := convert.Convert(index.Key, cty.String)
		if err != nil {
			kc = cty.StringVal("0")
		}

		k := kc.AsString()

		if vv, exists := ob[k]; exists {
			val := buildObject(traversal, vv.AsValueMap(), mock, i+1)
			ob[k] = cty.ObjectVal(val)
			return ob
		}

		val := buildObject(traversal, make(map[string]cty.Value), mock, i+1)
		ob[k] = cty.ObjectVal(val)
		return ob
	}

	if v, ok := traverser.(hcl.TraverseAttr); ok {
		if len(traversal)-1 == i {
			// if the attribute already exists, and we're not setting a list value
			// then we should return here. It's most likely that we weren't able to
			// get the full variable calls for the context, so resetting the value could
			// be harmful.
			if _, exists := ob[v.Name]; exists && mock.Type() == cty.String {
				return ob
			}

			ob[v.Name] = mock
			return ob
		}

		if vv, exists := ob[v.Name]; exists {
			if isList(vv) {
				items := make([]cty.Value, vv.LengthInt())
				it := vv.ElementIterator()
				for it.Next() {
					key, sourceItem := it.Element()
					val := buildObject(traversal, sourceItem.AsValueMap(), mock, i+1)
					i, _ := key.AsBigFloat().Int64()
					items[i] = cty.ObjectVal(val)
				}
				ob[v.Name] = cty.TupleVal(items)
				return ob
			}

			next := traversal[i+1]
			if _, ok := next.(hcl.TraverseIndex); ok {
				if !isList(vv) {
					vv = cty.TupleVal([]cty.Value{vv})
				}
			}

			val := buildObject(traversal, vv.AsValueMap(), mock, i+1)
			ob[v.Name] = cty.ObjectVal(val)
			return ob
		}

		val := buildObject(traversal, make(map[string]cty.Value), mock, i+1)
		ob[v.Name] = cty.ObjectVal(val)
		return ob
	}

	return buildObject(traversal, ob, mock, i+1)
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
func (attr *Attribute) Equals(val interface{}) bool {
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
			attr.Logger.Debugf("Error converting number for equality check. %s", err)
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
			attr.Logger.WithError(err).Warn("could not unpack int from block index attr, returning 0")
			return "0"
		}

		return fmt.Sprintf("%d", intVal)
	default:
		attr.Logger.Debugf("could not get index value for unsupported cty type %s, returning 0", part.Key.Type())
		return "0"
	}
}

// Reference returns the pointer to a Reference struct that holds information about the Attributes
// referenced block. Reference achieves this by traversing the Attribute Expression in order to find the
// parent block. E.g. with the following HCL
//
// 		resource "aws_launch_template" "foo2" {
// 			name = "foo2"
// 		}
//
//		resource "some_resource" "example_with_launch_template_3" {
//			...
//			name    = aws_launch_template.foo2.name
//		}
//
// The Attribute some_resource.name would have a reference of
//
//		Reference {
//			blockType: Type{
//				name:                  "resource",
//				removeTypeInReference: true,
//			}
//			typeLabel: "aws_launch_template"
//			nameLabel: "foo2"
//		}
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

func (attr *Attribute) referencesFromExpression(expression hcl.Expression) []*Reference {
	if attr == nil {
		return nil
	}

	countRef, _ := newReference([]string{"count", "index"})

	var refs []*Reference
	switch t := expression.(type) {
	case *hclsyntax.FunctionCallExpr:
		for _, arg := range t.Args {
			refs = append(refs, attr.referencesFromExpression(arg)...)
		}
	case *hclsyntax.ConditionalExpr:
		if ref, err := attr.createDotReferenceFromTraversal(t.TrueResult.Variables()...); err == nil {
			refs = append(refs, ref)
		}
		if ref, err := attr.createDotReferenceFromTraversal(t.FalseResult.Variables()...); err == nil {
			refs = append(refs, ref)
		}
		if ref, err := attr.createDotReferenceFromTraversal(t.Condition.Variables()...); err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.ScopeTraversalExpr:
		if ref, err := attr.createDotReferenceFromTraversal(t.Variables()...); err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.TemplateWrapExpr:
		refs = attr.referencesFromExpression(t.Wrapped)
	case *hclsyntax.TemplateExpr:
		for _, part := range t.Parts {
			ref, err := attr.createDotReferenceFromTraversal(part.Variables()...)
			if err != nil {
				continue
			}
			refs = append(refs, ref)
		}
	case *hclsyntax.TupleConsExpr:
		ref, err := attr.createDotReferenceFromTraversal(t.Variables()...)
		if err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.RelativeTraversalExpr:
		switch s := t.Source.(type) {
		case *hclsyntax.IndexExpr:
			if collectionRef, err := attr.createDotReferenceFromTraversal(s.Collection.Variables()...); err == nil {
				key, _ := s.Key.Value(attr.Ctx.Inner())
				collectionRef.SetKey(key)
				refs = append(refs, collectionRef, countRef)
			}
		default:
			if ref, err := attr.createDotReferenceFromTraversal(t.Source.Variables()...); err == nil {
				refs = append(refs, ref)
			}
		}
	case *hclsyntax.IndexExpr:
		if collectionRef, err := attr.createDotReferenceFromTraversal(t.Collection.Variables()...); err == nil {
			key, _ := t.Key.Value(attr.Ctx.Inner())
			collectionRef.SetKey(key)
			refs = append(refs, collectionRef, countRef)
		}
	default:
		attr.Logger.Debugf("could not create references for expression type: %s", t)
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
				}

				break
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
