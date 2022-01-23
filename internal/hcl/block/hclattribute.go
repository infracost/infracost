package block

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type Attribute interface {
	IsLiteral() bool
	Type() cty.Type
	Value() cty.Value
	Range() Range
	Name() string
	Contains(checkValue interface{}, equalityOptions ...EqualityOption) bool
	NotContains(checkValue interface{}, equalityOptions ...EqualityOption) bool
	HasIntersect(checkValues ...interface{}) bool
	StartsWith(prefix interface{}) bool
	EndsWith(suffix interface{}) bool
	Equals(checkValue interface{}, equalityOptions ...EqualityOption) bool
	NotEqual(checkValue interface{}, equalityOptions ...EqualityOption) bool
	RegexMatches(pattern interface{}) bool
	IsAny(options ...interface{}) bool
	IsNotAny(options ...interface{}) bool
	IsNone(options ...interface{}) bool
	IsTrue() bool
	IsFalse() bool
	IsEmpty() bool
	IsNotEmpty() bool
	IsNil() bool
	IsNotNil() bool
	MapValue(mapKey string) cty.Value
	LessThan(checkValue interface{}) bool
	LessThanOrEqualTo(checkValue interface{}) bool
	GreaterThan(checkValue interface{}) bool
	GreaterThanOrEqualTo(checkValue interface{}) bool
	IsDataBlockReference() bool
	Reference() (*Reference, error)
	AllReferences() []*Reference
	IsResourceBlockReference(resourceType string) bool
	ReferencesBlock(b Block) bool
	IsResolvable() bool
	IsNotResolvable() bool
	IsString() bool
	IsNumber() bool
	IsBool() bool
	ValueAsStrings() []string
	IsIterable() bool
	Each(f func(key cty.Value, val cty.Value))
}

type HCLAttribute struct {
	hclAttribute *hcl.Attribute
	ctx          *Context
}

func NewHCLAttribute(attr *hcl.Attribute, ctx *Context) *HCLAttribute {
	return &HCLAttribute{
		hclAttribute: attr,
		ctx:          ctx,
	}
}

func (attr *HCLAttribute) IsLiteral() bool {
	if attr == nil {
		return false
	}
	return len(attr.hclAttribute.Expr.Variables()) == 0
}

func (attr *HCLAttribute) IsResolvable() bool {
	if attr == nil {
		return false
	}
	return attr.Value() != cty.NilVal && attr.Value().IsKnown()
}

func (attr *HCLAttribute) IsNotResolvable() bool {
	return !attr.IsResolvable()
}

func (attr *HCLAttribute) Type() cty.Type {
	if attr == nil {
		return cty.NilType
	}
	return attr.Value().Type()
}

func (attr *HCLAttribute) IsIterable() bool {
	if attr == nil {
		return false
	}
	return attr.Value().Type().IsCollectionType() || attr.Value().Type().IsObjectType() || attr.Value().Type().IsMapType() || attr.Value().Type().IsListType() || attr.Value().Type().IsSetType() || attr.Value().Type().IsTupleType()
}

func (attr *HCLAttribute) Each(f func(key cty.Value, val cty.Value)) {
	if attr == nil {
		return
	}
	val := attr.Value()
	val.ForEachElement(func(key cty.Value, val cty.Value) (stop bool) {
		f(key, val)
		return false
	})
}

func (attr *HCLAttribute) IsString() bool {
	if attr == nil {
		return false
	}
	return !attr.Value().IsNull() && attr.Value().IsKnown() && attr.Value().Type() == cty.String
}

func (attr *HCLAttribute) IsNumber() bool {
	if attr == nil {
		return false
	}
	return !attr.Value().IsNull() && attr.Value().IsKnown() && attr.Value().Type() == cty.Number
}

func (attr *HCLAttribute) IsBool() bool {
	if attr == nil {
		return false
	}
	return !attr.Value().IsNull() && attr.Value().IsKnown() && attr.Value().Type() == cty.Bool
}

func (attr *HCLAttribute) Value() (ctyVal cty.Value) {
	if attr == nil {
		return cty.NilVal
	}
	defer func() {
		if err := recover(); err != nil {
			ctyVal = cty.NilVal
		}
	}()
	ctyVal, _ = attr.hclAttribute.Expr.Value(attr.ctx.Inner())
	if !ctyVal.IsKnown() {
		return cty.NilVal
	}
	return ctyVal
}

func (attr *HCLAttribute) Range() Range {
	if attr == nil {
		return Range{}
	}
	return Range{
		Filename:  attr.hclAttribute.Range.Filename,
		StartLine: attr.hclAttribute.Range.Start.Line,
		EndLine:   attr.hclAttribute.Range.End.Line,
	}
}

func (attr *HCLAttribute) Name() string {
	if attr == nil {
		return ""
	}
	return attr.hclAttribute.Name
}

func (attr *HCLAttribute) ValueAsStrings() (strings []string) {
	if attr == nil {
		return strings
	}
	defer func() {
		if err := recover(); err != nil {
			strings = nil
		}
	}()
	strings = getStrings(attr.hclAttribute.Expr, attr.ctx.Inner())
	return
}

func getStrings(expr hcl.Expression, ctx *hcl.EvalContext) []string {
	var results []string
	switch t := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		for _, expr := range t.Exprs {
			results = append(results, getStrings(expr, ctx)...)
		}
	case *hclsyntax.FunctionCallExpr, *hclsyntax.ScopeTraversalExpr,
		*hclsyntax.ConditionalExpr:
		subVal, err := t.Value(ctx)
		if err == nil && subVal.Type() == cty.String {
			results = append(results, subVal.AsString())
		}
	case *hclsyntax.LiteralValueExpr:
		if t.Val.Type() == cty.String {
			results = append(results, t.Val.AsString())
		}
	case *hclsyntax.TemplateExpr:
		// walk the parts of the expression to ensure that it has a literal value
		for _, p := range t.Parts {
			results = append(results, getStrings(p, ctx)...)
		}
	}
	return results
}

func (attr *HCLAttribute) listContains(val cty.Value, stringToLookFor string, ignoreCase bool) bool {
	if attr == nil {
		return false
	}
	valueSlice := val.AsValueSlice()
	for _, value := range valueSlice {
		stringToTest := value
		if value.Type().IsObjectType() || value.Type().IsMapType() {
			valueMap := value.AsValueMap()
			stringToTest = valueMap["key"]
		}
		if value.Type().HasDynamicTypes() {
			// References without a value can't logically "contain" a some string to check against.
			return false
		}
		if !value.IsKnown() {
			continue
		}
		if ignoreCase && strings.EqualFold(stringToTest.AsString(), stringToLookFor) {
			return true
		}
		if stringToTest.AsString() == stringToLookFor {
			return true
		}
	}
	return false
}

func (attr *HCLAttribute) mapContains(checkValue interface{}, val cty.Value) bool {
	if attr == nil {
		return false
	}
	valueMap := val.AsValueMap()
	switch t := checkValue.(type) {
	case map[interface{}]interface{}:
		for k, v := range t {
			for key, value := range valueMap {
				rawValue := getRawValue(value)
				if key == k && evaluate(v, rawValue) {
					return true
				}
			}
		}
		return false
	case map[string]interface{}:
		for k, v := range t {
			for key, value := range valueMap {
				rawValue := getRawValue(value)
				if key == k && evaluate(v, rawValue) {
					return true
				}
			}
		}
		return false
	default:
		for key := range valueMap {
			if key == checkValue {
				return true
			}
		}
		return false
	}
}

func (attr *HCLAttribute) NotContains(checkValue interface{}, equalityOptions ...EqualityOption) bool {
	return !attr.Contains(checkValue, equalityOptions...)
}

func (attr *HCLAttribute) Contains(checkValue interface{}, equalityOptions ...EqualityOption) bool {
	if attr == nil {
		return false
	}
	ignoreCase := false
	for _, option := range equalityOptions {
		if option == IgnoreCase {
			ignoreCase = true
		}
	}
	val := attr.Value()
	if val.IsNull() {
		return false
	}

	if val.Type().IsObjectType() || val.Type().IsMapType() {
		return attr.mapContains(checkValue, val)
	}

	stringToLookFor := fmt.Sprintf("%v", checkValue)

	if val.Type().IsListType() || val.Type().IsTupleType() {
		return attr.listContains(val, stringToLookFor, ignoreCase)
	}

	if ignoreCase && containsIgnoreCase(val.AsString(), stringToLookFor) {
		return true
	}

	return strings.Contains(val.AsString(), stringToLookFor)
}

func containsIgnoreCase(left, substring string) bool {
	return strings.Contains(strings.ToLower(left), strings.ToLower(substring))
}

func (attr *HCLAttribute) StartsWith(prefix interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.String {
		return strings.HasPrefix(attr.Value().AsString(), fmt.Sprintf("%v", prefix))
	}
	return false
}

func (attr *HCLAttribute) EndsWith(suffix interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.String {
		return strings.HasSuffix(attr.Value().AsString(), fmt.Sprintf("%v", suffix))
	}
	return false
}

type EqualityOption int

const (
	IgnoreCase EqualityOption = iota
)

func (attr *HCLAttribute) Equals(checkValue interface{}, equalityOptions ...EqualityOption) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.String {
		for _, option := range equalityOptions {
			if option == IgnoreCase {
				return strings.EqualFold(strings.ToLower(attr.Value().AsString()), strings.ToLower(fmt.Sprintf("%v", checkValue)))
			}
		}
		result := strings.EqualFold(attr.Value().AsString(), fmt.Sprintf("%v", checkValue))
		return result
	}
	if attr.Value().Type() == cty.Bool {
		return attr.Value().True() == checkValue
	}
	if attr.Value().Type() == cty.Number {
		checkNumber, err := gocty.ToCtyValue(checkValue, cty.Number)
		if err != nil {
			log.Debugf("Error converting number for equality check. %s", err)
			return false
		}
		return attr.Value().RawEquals(checkNumber)
	}

	return false
}

func (attr *HCLAttribute) NotEqual(checkValue interface{}, equalityOptions ...EqualityOption) bool {
	return !attr.Equals(checkValue, equalityOptions...)
}

func (attr *HCLAttribute) RegexMatches(pattern interface{}) bool {
	if attr == nil {
		return false
	}
	patternVal := fmt.Sprintf("%v", pattern)
	re, err := regexp.Compile(patternVal)
	if err != nil {
		log.Debugf("an error occurred while compiling the regex: %s", err)
		return false
	}
	if attr.Value().Type() == cty.String {
		match := re.MatchString(attr.Value().AsString())
		return match
	}
	return false
}

func (attr *HCLAttribute) IsNotAny(options ...interface{}) bool {
	return !attr.IsAny(options...)
}

func (attr *HCLAttribute) IsAny(options ...interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.String {
		value := attr.Value().AsString()
		for _, option := range options {
			if option == value {
				return true
			}
		}
	}
	if attr.Value().Type() == cty.Number {
		for _, option := range options {
			checkValue, err := gocty.ToCtyValue(option, cty.Number)
			if err != nil {
				log.Debugf("Error converting number for equality check. %s", err)
				return false
			}
			if attr.Value().RawEquals(checkValue) {
				return true
			}
		}
	}
	if attr.IsIterable() {
		attrVals := attr.Value().AsValueSlice()
		for _, option := range options {
			for _, attrVal := range attrVals {
				if attrVal.Type() == cty.String && attrVal.AsString() == option {
					return true
				}
			}
		}
	}
	return false
}

func (attr *HCLAttribute) IsNone(options ...interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.String {
		for _, option := range options {
			if option == attr.Value().AsString() {
				return false
			}
		}
	}
	if attr.Value().Type() == cty.Number {
		for _, option := range options {
			checkValue, err := gocty.ToCtyValue(option, cty.Number)
			if err != nil {
				log.Debugf("Error converting number for equality check. %s", err)
				return false
			}
			if attr.Value().RawEquals(checkValue) {
				return false
			}

		}
	}

	return true
}

func (attr *HCLAttribute) IsTrue() bool {
	if attr == nil {
		return false
	}
	switch attr.Value().Type() {
	case cty.Bool:
		return attr.Value().True()
	case cty.String:
		val := attr.Value().AsString()
		val = strings.Trim(val, "\"")
		return strings.ToLower(val) == "true"
	case cty.Number:
		val := attr.Value().AsBigFloat()
		f, _ := val.Float64()
		return f > 0
	}
	return false
}

func (attr *HCLAttribute) IsFalse() bool {
	if attr == nil {
		return false
	}
	switch attr.Value().Type() {
	case cty.Bool:
		return attr.Value().False()
	case cty.String:
		val := attr.Value().AsString()
		val = strings.Trim(val, "\"")
		return strings.ToLower(val) == "false"
	}
	return false
}

func (attr *HCLAttribute) IsEmpty() bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.String {
		return len(attr.Value().AsString()) == 0
	}
	if attr.Type().IsListType() || attr.Type().IsTupleType() {
		return len(attr.Value().AsValueSlice()) == 0
	}
	if attr.Type().IsMapType() || attr.Type().IsObjectType() {
		return len(attr.Value().AsValueMap()) == 0
	}
	if attr.Value().Type() == cty.Number {
		// a number can't ever be empty
		return false
	}
	if attr.Value().IsNull() {
		return attr.isNullAttributeEmpty()
	}
	return true
}

func (attr *HCLAttribute) IsNotEmpty() bool {
	return !attr.IsEmpty()
}

func (attr *HCLAttribute) isNullAttributeEmpty() bool {
	if attr == nil {
		return false
	}
	switch t := attr.hclAttribute.Expr.(type) {
	case *hclsyntax.FunctionCallExpr, *hclsyntax.ScopeTraversalExpr,
		*hclsyntax.ConditionalExpr, *hclsyntax.LiteralValueExpr:
		return false
	case *hclsyntax.TemplateExpr:
		// walk the parts of the expression to ensure that it has a literal value
		for _, p := range t.Parts {
			switch pt := p.(type) {
			case *hclsyntax.LiteralValueExpr:
				if pt != nil && !pt.Val.IsNull() {
					return false
				}
			case *hclsyntax.ScopeTraversalExpr:
				return false
			}
		}
	}
	return true
}

func (attr *HCLAttribute) MapValue(mapKey string) cty.Value {
	if attr == nil {
		return cty.NilVal
	}
	if attr.Type().IsObjectType() || attr.Type().IsMapType() {
		attrMap := attr.Value().AsValueMap()
		for key, value := range attrMap {
			if key == mapKey {
				return value
			}
		}
	}
	return cty.NilVal
}

func (attr *HCLAttribute) LessThan(checkValue interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.Number {
		checkNumber, err := gocty.ToCtyValue(checkValue, cty.Number)
		if err != nil {
			log.Debugf("Error converting number for equality check. %s", err)
			return false
		}

		return attr.Value().LessThan(checkNumber).True()
	}
	return false
}

func (attr *HCLAttribute) LessThanOrEqualTo(checkValue interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.Number {
		checkNumber, err := gocty.ToCtyValue(checkValue, cty.Number)
		if err != nil {
			log.Debugf("Error converting number for equality check. %s", err)
			return false
		}

		return attr.Value().LessThanOrEqualTo(checkNumber).True()
	}
	return false
}

func (attr *HCLAttribute) GreaterThan(checkValue interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.Number {
		checkNumber, err := gocty.ToCtyValue(checkValue, cty.Number)
		if err != nil {
			log.Debugf("Error converting number for equality check. %s", err)
			return false
		}

		return attr.Value().GreaterThan(checkNumber).True()
	}
	return false
}

func (attr *HCLAttribute) GreaterThanOrEqualTo(checkValue interface{}) bool {
	if attr == nil {
		return false
	}
	if attr.Value().Type() == cty.Number {
		checkNumber, err := gocty.ToCtyValue(checkValue, cty.Number)
		if err != nil {
			log.Debugf("Error converting number for equality check. %s", err)
			return false
		}

		return attr.Value().GreaterThanOrEqualTo(checkNumber).True()
	}
	return false
}

func (attr *HCLAttribute) IsDataBlockReference() bool {
	if attr == nil {
		return false
	}

	if t, ok := attr.hclAttribute.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
		split := t.Traversal.SimpleSplit()
		return split.Abs.RootName() == "data"
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

func (attr *HCLAttribute) Reference() (*Reference, error) {
	if attr == nil {
		return nil, fmt.Errorf("attribute is nil")
	}
	switch t := attr.hclAttribute.Expr.(type) {
	case *hclsyntax.RelativeTraversalExpr:
		switch s := t.Source.(type) {
		case *hclsyntax.IndexExpr:
			collectionRef, err := createDotReferenceFromTraversal(s.Collection.Variables()...)
			if err != nil {
				return nil, err
			}
			key, _ := s.Key.Value(attr.ctx.Inner())
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

func (attr *HCLAttribute) AllReferences() []*Reference {
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

func (attr *HCLAttribute) referencesInTemplate() []*Reference {
	if attr == nil {
		return nil
	}
	var refs []*Reference
	switch t := attr.hclAttribute.Expr.(type) {
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

func (attr *HCLAttribute) referencesInConditional() []*Reference {
	if attr == nil {
		return nil
	}
	var refs []*Reference
	if t, ok := attr.hclAttribute.Expr.(*hclsyntax.ConditionalExpr); ok {
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

func (attr *HCLAttribute) IsResourceBlockReference(resourceType string) bool {
	if attr == nil {
		return false
	}
	if t, ok := attr.hclAttribute.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
		split := t.Traversal.SimpleSplit()
		return split.Abs.RootName() == resourceType
	}
	return false
}

func (attr *HCLAttribute) ReferencesBlock(b Block) bool {
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

func getRawValue(value cty.Value) interface{} {
	typeName := value.Type().FriendlyName()

	switch typeName {
	case "string":
		return value.AsString()
	case "number":
		return value.AsBigFloat()
	case "bool":
		return value.True()
	}

	return value
}

func (attr *HCLAttribute) IsNil() bool {
	return attr == nil
}

func (attr *HCLAttribute) IsNotNil() bool {
	return !attr.IsNil()
}

func (attr *HCLAttribute) HasIntersect(checkValues ...interface{}) bool {
	if !attr.Type().IsListType() && !attr.Type().IsTupleType() {
		return false
	}

	for _, item := range checkValues {
		if attr.Contains(item) {
			return true
		}
	}
	return false

}
