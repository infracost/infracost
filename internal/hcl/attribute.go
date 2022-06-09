package hcl

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
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
}

// IsIterable returns if the attribute can be ranged over.
func (attr *Attribute) IsIterable() bool {
	if attr == nil {
		return false
	}

	return attr.Value().Type().IsCollectionType() || attr.Value().Type().IsObjectType() || attr.Value().Type().IsMapType() || attr.Value().Type().IsListType() || attr.Value().Type().IsSetType() || attr.Value().Type().IsTupleType()
}

// Value returns the Attribute with the underlying hcl.Expression of the hcl.Attribute evaluated with
// the Attribute Context. This returns a cty.Value with the values filled from any variables or references
// that the Context carries.
func (attr *Attribute) Value() cty.Value {
	if attr == nil {
		return cty.NilVal
	}

	return attr.value(0)
}

func (attr *Attribute) value(retry int) (ctyVal cty.Value) {
	defer func() {
		if err := recover(); err != nil {
			log.Debugf("could not evaluate value for attr: %s err: %s\n%s", attr.Name(), err, debug.Stack())
			ctyVal = cty.NilVal
		}
	}()

	var diag hcl.Diagnostics
	ctyVal, diag = attr.HCLAttr.Expr.Value(attr.Ctx.Inner())
	if diag.HasErrors() {
		mockedVal := cty.StringVal(uuid.NewString())
		if retry > 2 {
			return mockedVal
		}

		ctx := attr.Ctx.Inner()
		for _, d := range diag {
			badVariables := d.Expression.Variables()

			// if the diagnostic summary indicates that we were the attribute we attempted to fetch is unsupported
			// this is likely from a Terraform attribute that is built from the provider. We then try and build
			// a mocked attribute so that the module evaluation isn't harmed.
			var shouldRetry bool
			if d.Summary == missingAttributeDiagnostic {
				shouldRetry = true
				for _, traversal := range badVariables {
					traverseVarAndSetCtx(ctx, traversal, mockedVal)
				}
			}

			if d.Summary == valueIsNonIterableDiagnostic {
				shouldRetry = true
				for _, traversal := range badVariables {
					// let's first try and find the actual value for this bad variable.
					// If it has an actual value let's use that to pass into the list.
					val, _ := traversal.TraverseAbs(ctx)
					if val == cty.NilVal {
						val = mockedVal
					}

					list := cty.ListVal([]cty.Value{val})
					traverseVarAndSetCtx(ctx, traversal, list)
				}
			}

			// now that we've built a mocked attribute on the global context let's try and retrieve the value once again.
			if shouldRetry {
				return attr.value(retry + 1)
			}
		}

		if attr.Verbose {
			log.Debugf("error diagnostic return from evaluating %s err: %s", attr.HCLAttr.Name, diag.Error())
		}
	}

	if !ctyVal.IsKnown() {
		return cty.NilVal
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
	if v, ok := traverser.(hcl.TraverseAttr); ok {
		if len(traversal)-1 == i {
			ob[v.Name] = mock
			return ob
		}

		if _, exists := ob[v.Name]; exists {
			val := buildObject(traversal, ob[v.Name].AsValueMap(), mock, i+1)
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
			log.Debugf("Error converting number for equality check. %s", err)
			return false
		}
		return attr.Value().RawEquals(checkNumber)
	}

	return false
}

func createDotReferenceFromTraversal(traversals ...hcl.Traversal) (*Reference, error) {
	var refParts []string

	for _, x := range traversals {
		for _, p := range x {
			switch part := p.(type) {
			case hcl.TraverseRoot:
				refParts = append(refParts, part.Name)
			case hcl.TraverseAttr:
				refParts = append(refParts, part.Name)
			case hcl.TraverseIndex:
				refParts[len(refParts)-1] = fmt.Sprintf("%s[%s]", refParts[len(refParts)-1], getIndexValue(part))
			}
		}
	}
	return newReference(refParts)
}

func getIndexValue(part hcl.TraverseIndex) string {
	switch part.Key.Type() {
	case cty.String:
		return fmt.Sprintf("%q", part.Key.AsString())
	case cty.Number:
		var intVal int
		if err := gocty.FromCtyValue(part.Key, &intVal); err != nil {
			log.Warn("could not unpack the int, returning 0")
			return "0"
		}
		return fmt.Sprintf("%d", intVal)
	default:
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
		if ref, err := createDotReferenceFromTraversal(t.TrueResult.Variables()...); err == nil {
			refs = append(refs, ref)
		}
		if ref, err := createDotReferenceFromTraversal(t.FalseResult.Variables()...); err == nil {
			refs = append(refs, ref)
		}
		if ref, err := createDotReferenceFromTraversal(t.Condition.Variables()...); err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.ScopeTraversalExpr:
		if ref, err := createDotReferenceFromTraversal(t.Variables()...); err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.TemplateWrapExpr:
		refs = attr.referencesFromExpression(t.Wrapped)
	case *hclsyntax.TemplateExpr:
		for _, part := range t.Parts {
			ref, err := createDotReferenceFromTraversal(part.Variables()...)
			if err != nil {
				continue
			}
			refs = append(refs, ref)
		}
	case *hclsyntax.TupleConsExpr:
		ref, err := createDotReferenceFromTraversal(t.Variables()...)
		if err == nil {
			refs = append(refs, ref)
		}
	case *hclsyntax.RelativeTraversalExpr:
		switch s := t.Source.(type) {
		case *hclsyntax.IndexExpr:
			if collectionRef, err := createDotReferenceFromTraversal(s.Collection.Variables()...); err == nil {
				key, _ := s.Key.Value(attr.Ctx.Inner())
				collectionRef.SetKey(key)
				refs = append(refs, collectionRef, countRef)
			}
		default:
			if ref, err := createDotReferenceFromTraversal(t.Source.Variables()...); err == nil {
				refs = append(refs, ref)
			}
		}
	case *hclsyntax.IndexExpr:
		if collectionRef, err := createDotReferenceFromTraversal(t.Collection.Variables()...); err == nil {
			key, _ := t.Key.Value(attr.Ctx.Inner())
			collectionRef.SetKey(key)
			refs = append(refs, collectionRef, countRef)
		}
	}
	return refs
}
