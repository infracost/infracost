package hcl

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

var terraformSchemaV012 = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "terraform",
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
		{
			Type: "locals",
		},
		{
			Type:       "output",
			LabelNames: []string{"name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
		},
	},
}

// Blocks is a helper type around a slice of blocks to provide easy access
// finding blocks of type.
type Blocks []*Block

// OfType returns Blocks of the given type t. See terraformSchemaV012 for a list of possible types to lookup.
func (blocks Blocks) OfType(t string) Blocks {
	var results []*Block

	for _, block := range blocks {
		if block.Type() == t {
			results = append(results, block)
		}
	}

	return results
}

// Block wraps a hcl.Block type with additional methods and context.
// A Block is a core piece of HCL schema and represents a set of data.
// Most importantly a Block directly corresponds to a schema.Resource.
//
// Blocks can represent a number of different types - see terraformSchemaV012 for a list
// of potential HCL blocks available.
//
// e.g. a type resource block could look like this in HCL:
//
// 		resource "aws_lb" "lb1" {
//   		load_balancer_type = "application"
// 		}
//
// A Block can also have a set number of child Blocks, these child Blocks in turn can also have children.
// Blocks are recursive. The following example is represents a resource Block with child Blocks:
//
//		resource "aws_instance" "t3_standard_cpuCredits" {
//		  	ami           = "fake_ami"
//  		instance_type = "t3.medium"
//
//			# child Block starts here
//  		credit_specification {
//    			cpu_credits = "standard"
//  		}
//		}
//
// See Attribute for more info about how the values of Blocks are evaluated with their Context and returned.
type Block struct {
	// hclBlock is the underlying hcl.Block that has been parsed from a hcl.File
	hclBlock *hcl.Block
	// context is a pointer to the contextual information that can be used when evaluating
	// the Block and its Attributes.
	context *Context
	// moduleBlock represents the parent module of this Block. If this is nil the Block is part of the root
	// module (the top level directory in a Terraform dir).
	moduleBlock *Block
	// expanded marks if the block has been cloned or duplicated as part of a foreach or count.
	expanded bool
	// cloneIndex represents the index of the parent that this Block has been cloned from
	cloneIndex int
	// childBlocks holds information about any child Blocks that the Block may have. This can be empty.
	// See Block docs for more information about child Blocks.
	childBlocks Blocks
}

// NewHCLBlock returns a Block with Context and child Blocks initialised.
func NewHCLBlock(hclBlock *hcl.Block, ctx *Context, moduleBlock *Block) *Block {
	if ctx == nil {
		ctx = NewContext(&hcl.EvalContext{}, nil)
	}

	var children Blocks
	if body, ok := hclBlock.Body.(*hclsyntax.Body); ok {
		for _, b := range body.Blocks {
			children = append(children, NewHCLBlock(b.AsHCLBlock(), ctx, moduleBlock))
		}

		if hclBlock.Type == "resource" || hclBlock.Type == "data" {
			// add commonly used identifiers to the block so that if it's referenced by other
			// blocks in context evaluation.
			if _, ok := body.Attributes["id"]; !ok {
				body.Attributes["id"] = newUniqueAttribute("id")
			}

			if _, ok := body.Attributes["arn"]; !ok {
				body.Attributes["arn"] = newUniqueAttribute("arn")
			}
		}

		return &Block{
			context:     ctx,
			hclBlock:    hclBlock,
			moduleBlock: moduleBlock,
			childBlocks: children,
		}
	}

	// if we can't get a *hclsyntax.Body from this block let's try and parse blocks from the root level schema.
	// This might be because the *hcl.Block represents a whole file contents.
	content, _, diag := hclBlock.Body.PartialContent(terraformSchemaV012)
	if diag != nil && diag.HasErrors() {
		log.Debugf("error loading partial content from hcl file %s", diag.Error())

		return &Block{
			context:     ctx,
			hclBlock:    hclBlock,
			moduleBlock: moduleBlock,
			childBlocks: children,
		}
	}

	for _, hb := range content.Blocks {
		children = append(children, NewHCLBlock(hb, ctx, moduleBlock))
	}

	return &Block{
		context:     ctx,
		hclBlock:    hclBlock,
		moduleBlock: moduleBlock,
		childBlocks: children,
	}
}

func newUniqueAttribute(name string) *hclsyntax.Attribute {
	return &hclsyntax.Attribute{
		Name: name,
		Expr: &hclsyntax.LiteralValueExpr{
			Val: cty.StringVal(uuid.NewString()),
		},
	}
}

// InjectBlock takes a block and appends it to the Blocks childBlocks with the Block's
// attributes set as contextual values on the child. In most cases this is because we've expanded
// the block into further Blocks as part of a count or for_each.
func (b *Block) InjectBlock(block *Block, name string) {
	block.hclBlock.Labels = []string{}
	block.hclBlock.Type = name

	for attrName, attr := range block.AttributesAsMap() {
		b.context.Root().SetByDot(attr.Value(), fmt.Sprintf("%s.%s.%s", b.Reference().String(), name, attrName))
	}

	b.childBlocks = append(b.childBlocks, block)
}

// IsCountExpanded returns if the Block has been expanded as part of a for_each or count evaluation.
func (b *Block) IsCountExpanded() bool {
	return b.expanded
}

// Clone creates a duplicate of the block and sets the returned Block's Context to include the
// index provided. This is primarily used when Blocks are expanded as part of a count evaluation.
func (b *Block) Clone(index cty.Value) *Block {
	var childCtx *Context
	if b.context != nil {
		childCtx = b.context.NewChild()
	} else {
		childCtx = NewContext(&hcl.EvalContext{}, nil)
	}

	cloneHCL := *b.hclBlock

	clone := NewHCLBlock(&cloneHCL, childCtx, b.moduleBlock)
	if len(clone.hclBlock.Labels) > 0 {
		position := len(clone.hclBlock.Labels) - 1
		labels := make([]string, len(clone.hclBlock.Labels))
		for i := 0; i < len(labels); i++ {
			labels[i] = clone.hclBlock.Labels[i]
		}
		if index.IsKnown() && !index.IsNull() {
			switch index.Type() {
			case cty.Number:
				f, _ := index.AsBigFloat().Float64()
				labels[position] = fmt.Sprintf("%s[%d]", clone.hclBlock.Labels[position], int(f))
			case cty.String:
				labels[position] = fmt.Sprintf("%s[%q]", clone.hclBlock.Labels[position], index.AsString())
			default:
				log.Debugf("Invalid key type in iterable: %#v", index.Type())
				labels[position] = fmt.Sprintf("%s[%#v]", clone.hclBlock.Labels[position], index)
			}
		} else {
			labels[position] = fmt.Sprintf("%s[%d]", clone.hclBlock.Labels[position], b.cloneIndex)
		}
		clone.hclBlock.Labels = labels
	}

	indexVal, _ := gocty.ToCtyValue(index, cty.Number)
	clone.context.SetByDot(indexVal, "count.index")
	clone.expanded = true
	b.cloneIndex++

	return clone
}

// SetContext sets the Block.context to the provided ctx. This ctx is also set on the child Blocks as
// a child Context. Meaning that it can be used in traversal evaluation when looking up Context variables.
func (b *Block) SetContext(ctx *Context) {
	b.context = ctx
	for _, block := range b.childBlocks {
		block.SetContext(ctx.NewChild())
	}
}

// HasModuleBlock returns is the Block as a module associated with it. If it doesn't this means
// that this Block is part of the root Module.
func (b *Block) HasModuleBlock() bool {
	if b == nil {
		return false
	}

	return b.moduleBlock != nil
}

// ModuleAddress returns the address of the module associated with this Block or "" if it is part of the root Module
func (b *Block) ModuleAddress() string {
	if b == nil || !b.HasModuleBlock() {
		return ""
	}

	return b.moduleBlock.FullName()
}

// ModuleName returns the name of the module associated with this Block or "" if it is part of the root Module
func (b *Block) ModuleName() string {
	if b == nil || !b.HasModuleBlock() {
		return ""
	}

	return b.moduleBlock.TypeLabel()
}

// ModuleSource returns the "source" attribute from the associated Module or "" if it is part of the root Module
func (b *Block) ModuleSource() string {
	if b == nil || !b.HasModuleBlock() {
		return ""
	}

	attr := b.moduleBlock.GetAttribute("source")
	if attr == nil {
		return ""
	}

	value := attr.Value()

	if value.Type() != cty.String {
		return ""
	}

	return value.AsString()
}

// Provider returns the provider by first checking if it is explicitly set as an attribute, if it is not
// the first word in the snake_case name of the type is returned.  E.g. the type 'aws_instance' would
// return provider 'aws'
func (b *Block) Provider() string {
	if b == nil {
		return ""
	}

	attr := b.GetAttribute("provider")
	if attr != nil {
		value := attr.Value()
		if value.Type() == cty.String {
			// An explicit provider is provided so use that
			return value.AsString()
		}
	}

	// there's no explicit provider so get the provider implied as the prefix from the type
	parts := strings.Split(b.TypeLabel(), "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// GetChildBlock returns the first child Block that has the name provided. e.g:
// If the current Block looks like such:
//
//		resource "aws_instance" "t3_standard_cpuCredits" {
//		  	ami           = "fake_ami"
//  		instance_type = "t3.medium"
//
//  		credit_specification {
//    			cpu_credits = "standard"
//  		}
//
//			ebs_block_device {
//				device_name = "xvdj"
//			}
//		}
//
// Then "credit_specification" &  "ebs_block_device" would be valid names that could be used to retrieve child Blocks.
func (b *Block) GetChildBlock(name string) *Block {
	var returnBlock *Block
	if b == nil || b.hclBlock == nil {
		return returnBlock
	}

	for _, child := range b.childBlocks {
		if child.Type() == name {
			return child
		}
	}
	return returnBlock
}

func (b *Block) HasChild(childElement string) bool {
	return b.GetAttribute(childElement) != nil || b.GetChildBlock(childElement) != nil
}

// Children returns all the child Blocks associated with this Block.
func (b *Block) Children() Blocks {
	if b == nil || b.hclBlock == nil {
		return nil
	}

	return b.childBlocks
}

// GetAttributes returns a list of Attribute for this Block. Attributes are key value specification on a given
// Block. For example take the following hcl:
//
//		resource "aws_instance" "t3_standard_cpuCredits" {
//		  	ami           = "fake_ami"
//  		instance_type = "t3.medium"
//
//  		credit_specification {
//    			cpu_credits = "standard"
//  		}
//		}
//
// ami & instance_type are the Attributes of this Block and credit_specification is a child Block.
func (b *Block) GetAttributes() []*Attribute {
	var results []*Attribute
	if b == nil || b.hclBlock == nil {
		return nil
	}

	for _, attr := range b.getHCLAttributes() {
		results = append(results, &Attribute{HCLAttr: attr, Ctx: b.context})
	}

	return results
}

// GetAttribute returns the given attribute with the provided name. It will return nil if the attribute is not found.
// If we take the following Block example:
//
//		resource "aws_instance" "t3_standard_cpuCredits" {
//		  	ami           = "fake_ami"
//  		instance_type = "t3.medium"
//
//  		credit_specification {
//    			cpu_credits = "standard"
//  		}
//		}
//
// ami & instance_type are both valid Attribute names that can be used to lookup Block Attributes.
func (b *Block) GetAttribute(name string) *Attribute {
	var attr *Attribute
	if b == nil || b.hclBlock == nil {
		return attr
	}

	for _, attr := range b.GetAttributes() {
		if attr.Name() == name {
			return attr
		}
	}

	return attr
}

// AttributesAsMap returns the Attributes of this block as a map with the
// attribute name as the key and the value as the Attribute.
func (b *Block) AttributesAsMap() map[string]*Attribute {
	attributes := make(map[string]*Attribute)

	for _, attr := range b.GetAttributes() {
		attributes[attr.Name()] = attr
	}

	return attributes
}

func (b *Block) getHCLAttributes() hcl.Attributes {
	switch body := b.hclBlock.Body.(type) {
	case *hclsyntax.Body:
		attributes := make(hcl.Attributes)
		for _, a := range body.Attributes {
			attributes[a.Name] = a.AsHCLAttribute()
		}
		return attributes
	default:
		_, body, diag := b.hclBlock.Body.PartialContent(terraformSchemaV012)
		if diag != nil {
			return nil
		}
		attrs, diag := body.JustAttributes()
		if diag != nil {
			return nil
		}
		return attrs
	}
}

// Values returns the Block as a cty.Value with all the Attributes evaluated with the Block Context.
// This means that any variables or references will be replaced by their actual value. For example:
//
//		variable "instance_type" {
//			default = "t3.medium"
//		}
//
//		resource "aws_instance" "t3_standard_cpucredits" {
//		  	ami           = "fake_ami"
//  		instance_type = var.instance_type
//		}
//
// Would evaluate to a cty.Value of type Object with the instance_type Attribute holding the value "t3.medium".
func (b *Block) Values() cty.Value {
	values := make(map[string]cty.Value)

	for _, attribute := range b.GetAttributes() {
		values[attribute.Name()] = attribute.Value()
	}

	return cty.ObjectVal(values)
}

// Reference returns a Reference to the given Block this can be used to when printing
// out full names of Blocks to stdout or a file.
func (b *Block) Reference() *Reference {
	var parts []string
	if b.Type() != "resource" {
		parts = append(parts, b.Type())
	}

	parts = append(parts, b.Labels()...)
	ref, _ := newReference(parts)
	return ref
}

// LocalName is the name relative to the current module
func (b *Block) LocalName() string {
	return b.Reference().String()
}

// FullName returns the fully qualified Reference name as it relates to the Blocks position in the
// entire Terraform config tree. This includes module name. e.g.
//
// The following resource residing in a module named "web_app":
//
//		resource "aws_instance" "t3_standard" {
//		  	ami           = "fake_ami"
//  		instance_type = var.instance_type
//		}
//
// Would have its FullName as module.web_app.aws_instance.t3_standard
// FullName is what Terraform uses in its JSON output file.
func (b *Block) FullName() string {
	if b.moduleBlock != nil {
		return fmt.Sprintf(
			"%s.%s",
			b.moduleBlock.FullName(),
			b.LocalName(),
		)
	}

	return b.LocalName()
}

func (b *Block) Type() string {
	return b.hclBlock.Type
}

func (b *Block) Labels() []string {
	return b.hclBlock.Labels
}

func (b *Block) Context() *Context {
	return b.context
}

func (b *Block) TypeLabel() string {
	if len(b.Labels()) > 0 {
		return b.Labels()[0]
	}
	return ""
}

func (b *Block) NameLabel() string {
	if len(b.Labels()) > 1 {
		return b.Labels()[1]
	}
	return ""
}

func (b *Block) Label() string {
	return strings.Join(b.hclBlock.Labels, ".")
}

func loadBlocksFromFile(file *hcl.File) (hcl.Blocks, error) {
	contents, diags := file.Body.Content(terraformSchemaV012)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}

	if contents == nil {
		return nil, fmt.Errorf("file contents is empty")
	}

	return contents.Blocks, nil
}
