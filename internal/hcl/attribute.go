package hcl

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Attribute provides a wrapper struct around hcl.Attribute it provides
// helper methods and functionality for common interactions with hcl.Attribute.
type Attribute struct {
	// HCLAttr is the underlying hcl.Attribute that the Attribute references.
	HCLAttr *hcl.Attribute
	// Ctx is the context that the Attribute should be evaluated against. This propagates
	// any references from variables into the attribute.
	Ctx *Context
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
		log.Debugf("error diagnostic return from evaluating %s err: %s", attr.HCLAttr.Name, diag.Error())
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
			log.Debug("could not unpack the int, returning 0")
			return "0"
		}
		return fmt.Sprintf("%d", intVal)
	default:
		return "0"
	}
}

func (attr *Attribute) Reference() (*Reference, error) {
	switch t := attr.HCLAttr.Expr.(type) {
	case *hclsyntax.RelativeTraversalExpr:
		switch s := t.Source.(type) {
		case *hclsyntax.IndexExpr:
			collectionRef, err := createDotReferenceFromTraversal(s.Collection.Variables()...)
			if err != nil {
				return nil, err
			}
			key, _ := s.Key.Value(attr.Ctx.Inner())
			collectionRef.SetKey(key)
			return collectionRef, nil
		default:
			return createDotReferenceFromTraversal(t.Source.Variables()...)
		}
	case *hclsyntax.ScopeTraversalExpr:
		return createDotReferenceFromTraversal(t.Traversal)
	case *hclsyntax.TemplateExpr:
		refs := attr.referencesInTemplate()
		if len(refs) == 0 {
			return nil, fmt.Errorf("no references in template")
		}
		return refs[0], nil
	default:
		return nil, fmt.Errorf("not a reference: no scope traversal")
	}
}

func (attr *Attribute) AllReferences() []*Reference {
	if attr == nil {
		return nil
	}

	var refs []*Reference
	refs = append(refs, attr.referencesInTemplate()...)
	refs = append(refs, attr.referencesInConditional()...)
	ref, err := attr.Reference()
	if err == nil {
		refs = append(refs, ref)
	}

	return refs
}

func (attr *Attribute) referencesInTemplate() []*Reference {
	if attr == nil {
		return nil
	}
	var refs []*Reference
	switch t := attr.HCLAttr.Expr.(type) {
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
	}
	return refs
}

func (attr *Attribute) referencesInConditional() []*Reference {
	if attr == nil {
		return nil
	}

	var refs []*Reference
	if t, ok := attr.HCLAttr.Expr.(*hclsyntax.ConditionalExpr); ok {
		if ref, err := createDotReferenceFromTraversal(t.TrueResult.Variables()...); err == nil {
			refs = append(refs, ref)
		}
		if ref, err := createDotReferenceFromTraversal(t.FalseResult.Variables()...); err == nil {
			refs = append(refs, ref)
		}
		if ref, err := createDotReferenceFromTraversal(t.Condition.Variables()...); err == nil {
			refs = append(refs, ref)
		}
	}

	return refs
}

func (attr *Attribute) ReferencesBlock(b *Block) bool {
	if attr == nil {
		return false
	}

	for _, ref := range attr.AllReferences() {
		if ref.RefersTo(b) {
			return true
		}
	}

	return false
}

func (attr *Attribute) IsNil() bool {
	return attr == nil
}

func (attr *Attribute) IsNotNil() bool {
	return !attr.IsNil()
}
