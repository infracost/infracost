package hcl

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
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
