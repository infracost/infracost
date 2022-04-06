package hcl

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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
func (attr *Attribute) Value() (ctyVal cty.Value) {
	defer func() {
		if err := recover(); err != nil {
			ctyVal = cty.NilVal
		}
	}()

	var diag hcl.Diagnostics
	ctyVal, diag = attr.HCLAttr.Expr.Value(attr.Ctx.Inner())
	if diag.HasErrors() {
		if attr.Verbose {
			log.Debugf("error diagnostic return from evaluating %s err: %s", attr.HCLAttr.Name, diag.Error())
		}
	}

	if !ctyVal.IsKnown() {
		return cty.NilVal
	}

	return ctyVal
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
	var refs []*Reference
	switch t := expression.(type) {
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
				refs = append(refs, collectionRef)
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
			refs = append(refs, collectionRef)
		}
	}
	return refs
}
