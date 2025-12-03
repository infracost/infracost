package hcl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/hcl/funcs"
	"github.com/infracost/infracost/internal/hcl/mock"
	"github.com/infracost/infracost/internal/hcl/modules"
)

var (
	terraformSchemaV012 = &hcl.BodySchema{
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
	terraformAndProviderBlocks = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "terraform",
			},
			{
				Type:       "provider",
				LabelNames: []string{"name"},
			},
		},
	}
	nestedBackendBlock = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "backend",
				LabelNames: []string{"name"},
			},
		},
	}
	justModuleBlocks = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "module",
				LabelNames: []string{"name"},
			},
		},
	}

	errorNoHCLContents = fmt.Errorf("file contents is empty")
)

// referencedBlocks is a helper in interface adheres to the sort.Interface interface.
// This enables us to sort the blocks by their references to provide a list order
// safe for context evaluation.
type referencedBlocks []*Block

func (b referencedBlocks) Len() int      { return len(b) }
func (b referencedBlocks) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Less reports whether the Block with index i must sort before the Block with index j.
// If the i's name is referenced by Block j then i should start before j. This is
// because we need to evaluate the output of Block i before we can continue to j.
func (b referencedBlocks) Less(i, j int) bool {
	moduleName := b[i].Reference().nameLabel

	attrs := b[j].GetAttributes()
	for _, attr := range attrs {
		refs := attr.AllReferences()
		for _, ref := range refs {
			if ref.typeLabel == moduleName {
				return true
			}
		}
	}

	return false
}

// ModuleBlocks is a wrapper around SortedByCaller that selects just Modules to be sorted.
func (blocks Blocks) ModuleBlocks() Blocks {
	justModules := blocks.OfType("module")
	return justModules.SortedByCaller()
}

// SortedByCaller returns all the Blocks of type module. The returned Blocks
// are sorted in order of reference. Blocks that are referenced by others are
// the first in this list.
//
// So if we start with a list of [A,B,C] and A references B the returned list will
// be [B,A,C].
//
// This makes the list returned safe for context evaluation, as we evaluate modules that have
// outputs that other modules rely on first.
func (blocks Blocks) SortedByCaller() Blocks {
	sorted := make(Blocks, len(blocks))
	toSort := make(referencedBlocks, len(blocks))

	copy(toSort, blocks)
	sort.Sort(toSort)
	copy(sorted, toSort)

	return sorted
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

func (blocks Blocks) FindLocalName(name string) *Block {
	for _, block := range blocks {
		if block.LocalName() == name {
			return block
		}
	}

	return nil
}

// BlockMatcher defines a struct that can be used to filter a list of blocks to a single Block.
type BlockMatcher struct {
	Type       string
	Label      string
	StripCount bool
}

// Matching returns a single block filtered from the given pattern.
// If more than one Block is filtered by the pattern, Matching returns the first Block found.
func (blocks Blocks) Matching(pattern BlockMatcher) *Block {
	search := blocks

	if pattern.Type != "" {
		search = blocks.OfType(pattern.Type)
	}

	for _, block := range search {
		label := block.Label()
		if pattern.StripCount {
			label = stripCount(label)
		}

		if pattern.Label == label {
			return block
		}
	}

	if len(search) > 0 {
		return search[0]
	}

	return nil
}

// Outputs returns a map of all the evaluated outputs from the list of Blocks.
func (blocks Blocks) Outputs(suppressNil bool) cty.Value {
	data := make(map[string]cty.Value)

	for _, block := range blocks.OfType("output") {
		attr := block.GetAttribute("value")
		if attr == nil {
			continue
		}

		value := attr.Value()

		if suppressNil {
			// resolve the attribute value. This will evaluate any expressions that
			// the attribute uses and try and return the final value. If the end
			// value can't be resolved we set it as a blank string. This is
			// safe to use for callers and won't cause panics when marshalling
			// the returned cty.Value.
			if value == cty.NilVal {
				value = cty.StringVal("")
			}
		}

		data[block.Label()] = value
	}

	return cty.ObjectVal(data)
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
//			resource "aws_lb" "lb1" {
//	  		load_balancer_type = "application"
//			}
//
// A Block can also have a set number of child Blocks, these child Blocks in turn can also have children.
// Blocks are recursive. The following example is represents a resource Block with child Blocks:
//
//			resource "aws_instance" "t3_standard_cpuCredits" {
//			  	ami           = "fake_ami"
//	 		instance_type = "t3.medium"
//
//				# child Block starts here
//	 		credit_specification {
//	   			cpu_credits = "standard"
//	 		}
//			}
//
// See Attribute for more info about how the values of Blocks are evaluated with their Context and returned.
type Block struct {
	// HCLBlock is the underlying hcl.Block that has been parsed from a hcl.File
	HCLBlock *hcl.Block
	// UniqueAttrs specifies infracost specific unique attributes that will be appended to the block values.
	UniqueAttrs map[string]*hcl.Attribute
	// context is a pointer to the contextual information that can be used when evaluating
	// the Block and its Attributes.
	context *Context
	// moduleBlock represents the parent module of this Block. If this is nil the Block is part of the root
	// module (the top level directory in a Terraform dir).
	moduleBlock *Block
	// rootPath is the working directory of the CLI process.
	rootPath string
	// expanded marks if the block has been cloned or duplicated as part of a foreach or count.
	expanded bool
	// cloneIndex represents the index of the parent that this Block has been cloned from
	cloneIndex int
	// original is the block that the block was cloned from
	original *Block
	// childBlocks holds information about any child Blocks that the Block may have. This can be empty.
	// See Block docs for more information about child Blocks.
	childBlocks Blocks
	// parent is the parent block if this is a child block.
	parent *Block
	// verbose determines whether the block uses verbose debug logging.
	verbose bool
	logger  zerolog.Logger
	// isGraph is a flag that indicates if the attribute should be evaluated with the graph evaluation
	isGraph    bool
	newMock    func(attr *Attribute) cty.Value
	attributes []*Attribute
	reference  *Reference

	Filename  string
	StartLine int
	EndLine   int
}

// BlockBuilder handles generating new Blocks as part of the parsing and evaluation process.
type BlockBuilder struct {
	MockFunc      func(a *Attribute) cty.Value
	SetAttributes []SetAttributesFunc
	Logger        zerolog.Logger
	HCLParser     *modules.SharedHCLParser
	isGraph       bool
}

// NewBlock returns a Block with Context and child Blocks initialised.
func (b BlockBuilder) NewBlock(filename string, rootPath string, hclBlock *hcl.Block, ctx *Context, parent *Block, moduleBlock *Block) *Block {
	if ctx == nil {
		ctx = NewContext(&hcl.EvalContext{}, nil, b.Logger)
	}

	// if the filepath is absolute let's make it relative to the working directory so
	// we trip any user/machine defined paths.
	if filepath.IsAbs(filename) {
		wd, _ := os.Getwd()
		rel, err := filepath.Rel(wd, filename)
		if err == nil {
			filename = rel
		}
	}

	isLoggingVerbose := strings.TrimSpace(os.Getenv("INFRACOST_HCL_DEBUG_VERBOSE")) == "true"
	if body, ok := hclBlock.Body.(*hclsyntax.Body); ok {
		block := &Block{
			Filename:    filename,
			StartLine:   body.SrcRange.Start.Line,
			EndLine:     body.SrcRange.End.Line,
			context:     ctx,
			UniqueAttrs: map[string]*hcl.Attribute{},
			HCLBlock:    hclBlock,
			moduleBlock: moduleBlock,
			rootPath:    rootPath,
			childBlocks: make(Blocks, len(body.Blocks)),
			verbose:     isLoggingVerbose,
			isGraph:     b.isGraph,
			newMock:     b.MockFunc,
			parent:      parent,
		}

		block.setLogger(b.Logger)

		for _, f := range b.SetAttributes {
			f(block)
		}

		for i, bb := range body.Blocks {
			block.childBlocks[i] = b.NewBlock(filename, rootPath, bb.AsHCLBlock(), ctx, block, moduleBlock)
		}

		return block
	}

	// if we can't get a *hclsyntax.Body from this block let's try and parse blocks from the root level schema.
	// This might be because the *hcl.Block represents a whole file contents.
	content, _, diag := hclBlock.Body.PartialContent(terraformSchemaV012)
	if diag != nil && diag.HasErrors() {
		b.Logger.Debug().Msgf("error loading partial content from hcl file %s", diag.Error())

		block := &Block{
			Filename:    filename,
			StartLine:   hclBlock.DefRange.Start.Line,
			EndLine:     hclBlock.DefRange.End.Line,
			context:     ctx,
			HCLBlock:    hclBlock,
			UniqueAttrs: map[string]*hcl.Attribute{},
			moduleBlock: moduleBlock,
			rootPath:    rootPath,
			verbose:     isLoggingVerbose,
			isGraph:     b.isGraph,
			newMock:     b.MockFunc,
		}
		block.setLogger(b.Logger)

		return block
	}

	block := &Block{
		Filename:    filename,
		StartLine:   hclBlock.DefRange.Start.Line,
		EndLine:     hclBlock.DefRange.End.Line,
		context:     ctx,
		HCLBlock:    hclBlock,
		UniqueAttrs: map[string]*hcl.Attribute{},
		moduleBlock: moduleBlock,
		rootPath:    rootPath,
		childBlocks: make(Blocks, len(content.Blocks)),
		verbose:     isLoggingVerbose,
		isGraph:     b.isGraph,
		newMock:     b.MockFunc,
	}

	for i, hb := range content.Blocks {
		block.childBlocks[i] = b.NewBlock(filename, rootPath, hb, ctx, block, moduleBlock)
	}

	block.setLogger(b.Logger)
	return block
}

// CloneBlock creates a duplicate of the block and sets the returned Block's Context to include the
// index provided. This is primarily used when Blocks are expanded as part of a count evaluation.
func (b BlockBuilder) CloneBlock(block *Block, index cty.Value) *Block {
	var childCtx *Context
	if block.context != nil {
		childCtx = block.context.NewChild()
	} else {
		childCtx = NewContext(&hcl.EvalContext{}, nil, b.Logger)
	}

	cloneHCL := *block.HCLBlock

	clone := b.NewBlock(block.Filename, block.rootPath, &cloneHCL, childCtx, block.parent, block.moduleBlock)
	if len(clone.HCLBlock.Labels) > 0 {
		position := len(clone.HCLBlock.Labels) - 1
		labels := make([]string, len(clone.HCLBlock.Labels))
		for i := range labels {
			labels[i] = clone.HCLBlock.Labels[i]
		}
		if index.IsKnown() && !index.IsNull() {
			switch index.Type() {
			case cty.Number:
				f, _ := index.AsBigFloat().Float64()
				labels[position] = fmt.Sprintf("%s[%d]", clone.HCLBlock.Labels[position], int(f))
			case cty.String:
				labels[position] = fmt.Sprintf("%s[%q]", clone.HCLBlock.Labels[position], index.AsString())
			default:
				b.Logger.Debug().Msgf("Invalid key type in iterable: %#v", index.Type())
				labels[position] = fmt.Sprintf("%s[%#v]", clone.HCLBlock.Labels[position], index)
			}
		} else {
			labels[position] = fmt.Sprintf("%s[%d]", clone.HCLBlock.Labels[position], block.cloneIndex)
		}
		clone.SetLabels(labels)
	}

	indexVal, _ := gocty.ToCtyValue(index, cty.Number)
	clone.context.SetByDot(indexVal, "count.index")
	clone.expanded = true
	block.cloneIndex++

	clone.original = block
	return clone
}

// BuildModuleBlocks loads all the Blocks for the module at the given path
func (b BlockBuilder) BuildModuleBlocks(block *Block, modulePath string, rootPath string) (Blocks, error) {
	var blocks Blocks
	moduleFiles, err := loadDirectory(b.HCLParser, b.Logger, modulePath, true)
	if err != nil {
		return blocks, fmt.Errorf("failed to load module %s: %w", block.Label(), err)
	}

	moduleCtx := NewContext(&hcl.EvalContext{}, nil, b.Logger)
	for _, file := range moduleFiles {
		fileBlocks, err := loadBlocksFromFile(file, nil)
		if err != nil {
			return blocks, err
		}

		if len(fileBlocks) > 0 {
			b.Logger.Debug().Msgf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}

		for _, fileBlock := range fileBlocks {
			blocks = append(blocks, b.NewBlock(file.path, rootPath, fileBlock, moduleCtx, nil, block))
		}
	}

	return blocks, err
}

// SetAttributesFunc defines a function that sets required attributes on a hcl.Block.
// This is done so that identifiers that are normally propagated from a Terraform state/apply
// are set on the Block. This means they can be used properly in references and outputs.
type SetAttributesFunc func(b *Block)

// SetUUIDAttributes adds commonly used identifiers to the block so that it can be referenced by other
// blocks in context evaluation. The identifiers are only set if they don't already exist as attributes
// on the block.
func SetUUIDAttributes(b *Block) {
	t := b.Type()
	if t != "resource" && t != "data" {
		return
	}

	if body, ok := b.HCLBlock.Body.(*hclsyntax.Body); ok {
		h := sha256.New()
		h.Write([]byte(b.FullName()))
		addressSha := hex.EncodeToString(h.Sum(nil))

		_, withCount := body.Attributes["count"]
		_, withEach := body.Attributes["for_each"]

		if _, ok := body.Attributes["id"]; !ok {
			b.UniqueAttrs["id"] = newUniqueAttribute(addressSha, "id", withCount, withEach)
		}

		if _, ok := body.Attributes["name"]; !ok {
			b.UniqueAttrs["name"] = newUniqueAttribute(addressSha, "name", withCount, withEach)
		}

		if _, ok := body.Attributes["arn"]; !ok {
			b.UniqueAttrs["arn"] = newArnAttribute(addressSha, "arn", withCount, withEach)
		}

		if _, ok := body.Attributes["self_link"]; !ok {
			b.UniqueAttrs["self_link"] = newUniqueAttribute(addressSha, "self_link", withCount, withEach)
		}
	}
}

func newUniqueAttribute(addressSha, name string, withCount bool, withEach bool) *hcl.Attribute {
	// prefix ids with hcl- so they can be identified as fake
	var exp hclsyntax.Expression = &hclsyntax.LiteralValueExpr{
		Val: cty.StringVal("hcl-" + addressSha),
	}

	if withCount {
		e, diags := hclsyntax.ParseExpression([]byte(`"hcl-`+addressSha+`-${count.index}"`), name, hcl.Pos{})
		if !diags.HasErrors() {
			exp = e
		}
	}

	if withEach {
		e, diags := hclsyntax.ParseExpression([]byte(`"hcl-`+addressSha+`-${each.key}"`), name, hcl.Pos{})
		if !diags.HasErrors() {
			exp = e
		}
	}

	return &hcl.Attribute{
		Name: name,
		Expr: exp,
	}
}

func newArnAttribute(addressSha, name string, withCount bool, withEach bool) *hcl.Attribute {

	// fakeARN replicates an aws arn string it deliberately leaves the
	// region section (in between the 3rd and 4th semicolon) blank as
	// Infracost will try and parse this region later down the line.
	// Keeping it blank will defer the region to what the provider has defined.
	fakeARN := fmt.Sprintf("arn:aws:hcl::%s", addressSha)
	var exp hclsyntax.Expression = &hclsyntax.LiteralValueExpr{
		Val: cty.StringVal(fakeARN),
	}

	if withCount {
		e, diags := hclsyntax.ParseExpression([]byte(`"`+fakeARN+`-${count.index}"`), name, hcl.Pos{})
		if !diags.HasErrors() {
			exp = e
		}
	}

	if withEach {
		e, diags := hclsyntax.ParseExpression([]byte(fakeARN+`-${each.key}"`), name, hcl.Pos{})
		if !diags.HasErrors() {
			exp = e
		}
	}

	return &hcl.Attribute{
		Name: name,
		Expr: exp,
	}
}

// IsCountExpanded returns if the Block has been expanded as part of a for_each or count evaluation.
func (b *Block) IsCountExpanded() bool {
	return b.expanded
}

// IsForEachReferencedExpanded checks if the block referenced under the for_each has already been expanded.
// This is used to check is we can safely expand this block, expanding block prematurely can lead to
// output inconsistencies. It is advised to always check if that the block has any references that are yet
// to be expanded before expanding itself.
func (b *Block) IsForEachReferencedExpanded(moduleBlocks Blocks) bool {
	attr := b.GetAttribute("for_each")
	if attr == nil {
		return true
	}

	r, err := attr.Reference()
	if err != nil || r == nil {
		return true
	}

	blockType := r.blockType.Name()
	if _, ok := validBlocksToExpand[blockType]; !ok {
		return true
	}

	label := r.String()
	if blockType == "module" {
		label = r.typeLabel
	}

	referenced := moduleBlocks.Matching(BlockMatcher{
		Type:       blockType,
		Label:      label,
		StripCount: true,
	})

	if referenced == nil {
		return true
	}

	return !referenced.ShouldExpand()
}

func (b *Block) ShouldExpand() bool {
	if b.IsCountExpanded() {
		return false
	}

	validType := b.Type() == "resource" || b.Type() == "module" || b.Type() == "data"
	if !validType {
		return false
	}

	countAttr := b.GetAttribute("count")
	forEachAttr := b.GetAttribute("for_each")

	return countAttr != nil || forEachAttr != nil
}

// SetContext sets the Block.context to the provided ctx. This ctx is also set on the child Blocks as
// a child Context. Meaning that it can be used in traversal evaluation when looking up Context variables.
func (b *Block) SetContext(ctx *Context) {
	b.context = ctx
	for _, attribute := range b.attributes {
		attribute.Ctx = ctx
	}

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

type ModuleMetadata struct {
	Filename  string `json:"filename"`
	BlockName string `json:"blockName"`
	StartLine int    `json:"startLine,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
}

func (b *Block) setLogger(logger zerolog.Logger) {
	// Use the provided logger as is initially so we avoid a nil pointer in the case where we need to log
	// an error while calculating b.FullName().
	b.logger = logger

	blockLogger := logger.With().Str("block_name", b.FullName()).Logger()

	b.logger = blockLogger
}

// CallDetails returns the tree of module calls that were used to create this resource. Each step of the tree
// contains a full file path and block name that were used to create the resource.
//
// CallDetails returns a list of ModuleMetadata that are ordered by appearance in the Terraform config tree.
func (b *Block) CallDetails() []ModuleMetadata {
	block := b
	var meta []ModuleMetadata
	for {
		meta = append(meta, ModuleMetadata{
			Filename:  block.Filename,
			BlockName: stripCount(block.LocalName()),
			StartLine: block.StartLine,
			EndLine:   block.EndLine,
		})

		if block.moduleBlock == nil {
			break
		}

		block = block.moduleBlock
	}

	reversed := make([]ModuleMetadata, 0, len(meta))
	for i := len(meta) - 1; i >= 0; i-- {
		reversed = append(reversed, meta[i])
	}

	return reversed
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

	return attr.AsString()
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
		value := attr.AsString()
		r, err := attr.Reference()
		if err == nil {
			// An explicit provider is provided so use that
			return r.String()
		}

		if value != "" {
			// An explicit provider is provided so use that
			return value
		}
	}

	// there's no explicit provider so get the provider implied as the prefix from the type
	parts := strings.Split(b.TypeLabel(), "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ProviderConfigKey looks up the key used to reference the provider in the "configuration.providers"
// section of the terraform plan json.  This should be used to set the "provider_config_key "
// of the resource in the "configuration.resources" section of plan json.
func (b *Block) ProviderConfigKey() string {
	provider := b.Provider()

	v := getFromProvider(b, provider, "config_key")
	var str string
	err := gocty.FromCtyValue(v, &str)
	if err != nil {
		// fall back to using the provider name
		return provider
	}
	return str
}

// GetChildBlock is a helper method around GetChildBlocks. It returns the first non nil child block matching name.
func (b *Block) GetChildBlock(name string) *Block {
	blocks := b.GetChildBlocks(name)
	if len(blocks) > 0 {
		return blocks[0]
	}

	return nil
}

// GetChildBlocks returns all the child Block that match the name provided. e.g:
// If the current Block looks like such:
//
//			resource "aws_instance" "t3_standard_cpuCredits" {
//			  	ami           = "fake_ami"
//	 		instance_type = "t3.medium"
//
//	 		credit_specification {
//	   			cpu_credits = "standard"
//	 		}
//
//				ebs_block_device {
//					device_name = "xvdj"
//				}
//			}
//
// Then "credit_specification" &  "ebs_block_device" would be valid names that could be used to retrieve child Blocks.
func (b *Block) GetChildBlocks(name string) []*Block {
	if b == nil || b.HCLBlock == nil {
		return nil
	}

	var children Blocks
	for _, child := range b.childBlocks {
		if child.Type() == name {
			children = append(children, child)
		}
	}

	return children
}

func (b *Block) HasChild(childElement string) bool {
	return b.GetAttribute(childElement) != nil || b.GetChildBlock(childElement) != nil
}

// Children returns all the child Blocks associated with this Block.
func (b *Block) Children() Blocks {
	if b == nil || b.HCLBlock == nil {
		return nil
	}

	children := make(Blocks, 0, len(b.childBlocks))
	for _, child := range b.childBlocks {
		// Skip lifecycle meta argument blocks since it never needs to be evaluated
		if supportsLifecycle(child) && child.Type() == "lifecycle" {
			continue
		}

		children = append(children, child)
	}

	return children
}

// GetAttributes returns a list of Attribute for this Block. Attributes are key value specification on a given
// Block. For example take the following hcl:
//
//			resource "aws_instance" "t3_standard_cpuCredits" {
//			  	ami           = "fake_ami"
//	 		instance_type = "t3.medium"
//
//	 		credit_specification {
//	   			cpu_credits = "standard"
//	 		}
//			}
//
// ami & instance_type are the Attributes of this Block and credit_specification is a child Block.
func (b *Block) GetAttributes() []*Attribute {
	if b == nil || b.HCLBlock == nil {
		return nil
	}

	if b.attributes != nil {
		return b.attributes
	}

	hclAttributes := b.getHCLAttributes()

	// Sort the attributes so the order is deterministic
	keys := make([]string, 0, len(hclAttributes)+len(b.UniqueAttrs))
	for k := range hclAttributes {
		keys = append(keys, k)
	}
	for k := range b.UniqueAttrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var attributes = make([]*Attribute, 0, len(keys))
	for _, k := range keys {
		if v, ok := b.UniqueAttrs[k]; ok {
			attributes = append(attributes, &Attribute{
				newMock: b.newMock,
				HCLAttr: v,
				Ctx:     b.context,
				Verbose: b.verbose,
				Logger: b.logger.With().Str(
					"attribute_name", k,
				).Logger(),
				IsGraph: b.isGraph,
			})
			continue
		}

		attributes = append(attributes, &Attribute{
			newMock: b.newMock,
			HCLAttr: hclAttributes[k],
			Ctx:     b.context,
			Verbose: b.verbose,
			Logger: b.logger.With().Str(
				"attribute_name", hclAttributes[k].Name,
			).Logger(),
			IsGraph: b.isGraph,
		})
	}

	if b.Type() == "data" && b.TypeLabel() == "local_file" {
		attributes = b.loadFileContentsToAttributes(attributes)
	}

	b.attributes = attributes
	return attributes
}

func (b *Block) loadFileContentsToAttributes(attributes []*Attribute) []*Attribute {
	for _, attribute := range attributes {
		if attribute.Name() == "filename" {
			content, err := funcs.File(b.rootPath, attribute.Value())
			if err != nil {
				b.logger.Debug().Err(err).Msgf("failed to load %s file contents", b.FullName())
				break
			}

			attributes = []*Attribute{
				attribute,
				b.syntheticAttribute("content", content),
			}

			break
		}
	}

	return attributes
}

func (b *Block) syntheticAttribute(name string, val cty.Value) *Attribute {
	rng := hcl.Range{
		Filename: b.Filename,
		Start:    hcl.Pos{Line: 1, Column: 1},
		End:      hcl.Pos{Line: 1, Column: 1},
	}

	hclAttr := &hcl.Attribute{
		Name: name,
		Expr: &hclsyntax.LiteralValueExpr{
			Val:      val,
			SrcRange: rng,
		},
		NameRange: rng,
		Range:     rng,
	}

	return &Attribute{
		newMock: b.newMock,
		HCLAttr: hclAttr,
		Ctx:     b.context,
		Verbose: b.verbose,
		Logger: b.logger.With().Str(
			"attribute_name", name,
		).Logger(),
		IsGraph: b.isGraph,
	}
}

// GetAttribute returns the given attribute with the provided name. It will return nil if the attribute is not found.
// If we take the following Block example:
//
//			resource "aws_instance" "t3_standard_cpuCredits" {
//			  	ami           = "fake_ami"
//	 		instance_type = "t3.medium"
//
//	 		credit_specification {
//	   			cpu_credits = "standard"
//	 		}
//			}
//
// ami & instance_type are both valid Attribute names that can be used to lookup Block Attributes.
func (b *Block) GetAttribute(name string) *Attribute {
	var attr *Attribute
	if b == nil || b.HCLBlock == nil {
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
	supportsDependsOn := supportsDependsOn(b)

	switch body := b.HCLBlock.Body.(type) {
	case *hclsyntax.Body:
		attributes := make(hcl.Attributes)
		for _, a := range body.Attributes {
			if _, ok := b.UniqueAttrs[a.Name]; ok {
				continue
			}

			// Ignore the depends_on meta attribute since it never needs to be evaluated
			if supportsDependsOn && a.Name == "depends_on" {
				continue
			}

			attributes[a.Name] = a.AsHCLAttribute()
		}
		return attributes
	default:
		_, body, diag := b.HCLBlock.Body.PartialContent(terraformSchemaV012)
		if diag != nil {
			return nil
		}
		attrs, diag := body.JustAttributes()
		if diag != nil {
			return nil
		}
		for k := range attrs {
			if _, ok := b.UniqueAttrs[k]; ok {
				delete(attrs, k)
			}

			// Ignore the depends_on meta attribute since it never needs to be evaluated
			if supportsDependsOn && k == "depends_on" {
				delete(attrs, k)
			}
		}

		return attrs
	}
}

// Values returns the Block as a cty.Value with all the Attributes evaluated with the Block Context.
// This means that any variables or references will be replaced by their actual value. For example:
//
//			variable "instance_type" {
//				default = "t3.medium"
//			}
//
//			resource "aws_instance" "t3_standard_cpucredits" {
//			  	ami           = "fake_ami"
//	 		instance_type = var.instance_type
//			}
//
// Would evaluate to a cty.Value of type Object with the instance_type Attribute holding the value "t3.medium".
func (b *Block) Values() cty.Value {
	if f, ok := blockValueFuncs[fmt.Sprintf("%s.%s", b.Type(), b.TypeLabel())]; ok {
		return f(b)
	}

	return b.values()
}

type AttributeWithUnknownKeys struct {
	Attribute        string   `json:"attribute"`
	MissingVariables []string `json:"missingVariables"`
}

func (b *Block) AttributesWithUnknownKeys() []AttributeWithUnknownKeys {
	var output []AttributeWithUnknownKeys
	for _, attr := range b.attributes {
		if causes := attr.ReferencesCausingUnknownKeys(); len(causes) > 0 {
			output = append(output, AttributeWithUnknownKeys{
				Attribute:        attr.Name(),
				MissingVariables: causes,
			})
		}
	}
	return output
}

func (b *Block) values() cty.Value {
	values := make(map[string]cty.Value)

	for _, attribute := range b.GetAttributes() {
		if attribute.Name() == "for_each" {
			continue
		}

		values[attribute.Name()] = attribute.Value()
	}

	// @TODO this needs to include all blocks in the future. However, we are limiting to just provider blocks just now
	// as the repercussions for this change could be quite vast. See https://github.com/infracost/infracost/issues/2596 for
	// more information.
	if b.Type() == "provider" {
		for _, child := range b.Children() {
			key := child.Type()
			if key == "dynamic" {
				continue
			}

			if v, ok := values[key]; ok {
				list := append(v.AsValueSlice(), child.values())
				if !cty.CanListVal(list) {
					b.logger.Debug().Msgf("ignoring child block %#v value with inconsistent list element types", key)
					continue
				}
				values[key] = cty.ListVal(list)
			} else {
				values[key] = cty.ListVal([]cty.Value{child.values()})
			}
		}

		// Add config_key as a value of the provider so resources can lookup the
		// correct provider_config_key using `getFromProvider(provider, "config_key")`
		configKey := b.Label()

		var alias string
		_ = gocty.FromCtyValue(values["alias"], &alias)
		if alias != "" {
			configKey = configKey + "." + alias
		}

		if b.ModuleAddress() != "" {
			configKey = b.ModuleAddress() + ":" + configKey
		}

		values["config_key"] = cty.StringVal(configKey)
	}

	return cty.ObjectVal(values)
}

// Reference returns a Reference to the given Block this can be used to when printing
// out full names of Blocks to stdout or a file.
func (b *Block) Reference() *Reference {
	if b.reference != nil && b.reference.String() != "" {
		return b.reference
	}

	var parts []string

	parent := b.parent
	for parent != nil {
		var parentParts []string

		if parent.Type() != "resource" {
			parentParts = append(parentParts, parent.Type())
		}

		parentParts = append(parentParts, parent.Labels()...)

		parts = append(parentParts, parts...)
		parent = parent.parent
	}

	if b.Type() != "resource" {
		parts = append(parts, b.Type())
	}

	parts = append(parts, b.Labels()...)
	ref, err := newReference(parts)
	if err != nil {
		b.logger.Debug().Err(err).Msgf(
			"returning empty block reference because we encountered an error generating a new reference",
		)
		ref = &Reference{}
	}

	b.reference = ref

	return b.reference
}

// VerticesReferenced traverses the block attributes and child blocks to build a complete
// list of all the vertices referenced by the Block. We build a special VertexReference
// if the block uses a provider but doesn't have one referenced. This is because blocks
// will depend on the default provider to get the correct region data, even if not explicitly
// referenced as a reference.
func (b *Block) VerticesReferenced() []VertexReference {
	var refs []VertexReference

	hasProviderAttr := false

	for _, attr := range b.GetAttributes() {
		if attr.Name() == "provider" {
			hasProviderAttr = true
		}

		refs = append(refs, attr.VerticesReferenced(b)...)
	}

	for _, childBlock := range b.Children() {
		refs = append(refs, childBlock.VerticesReferenced()...)
	}

	if !hasProviderAttr && usesProviderConfiguration(b) {
		providerName := b.Provider()
		if providerName != "" {
			refs = append(refs, VertexReference{
				Key: fmt.Sprintf("provider.%s", providerName),
			})
		}
	}

	return refs
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
//			resource "aws_instance" "t3_standard" {
//			  	ami           = "fake_ami"
//	 		instance_type = var.instance_type
//			}
//
// Would have its FullName as module.web_app.aws_instance.t3_standard
// FullName is what Terraform uses in its JSON output file.
func (b *Block) FullName() string {
	if b == nil {
		return ""
	}

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
	if b == nil || b.HCLBlock == nil {
		return ""
	}

	return b.HCLBlock.Type
}

func (b *Block) SetType(t string) {
	b.reference = nil
	b.HCLBlock.Type = t
}

func (b *Block) Labels() []string {
	if b == nil || b.HCLBlock == nil {
		return nil
	}

	return b.HCLBlock.Labels
}

func (b *Block) SetLabels(labels []string) {
	b.reference = nil
	b.HCLBlock.Labels = labels
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

// HasDynamicBlock searches all the nested children of the given block to see if
// any are type "dynamic". This is used before embarking on dynamic expansion
// logic.
func (b *Block) HasDynamicBlock() bool {
	if b == nil {
		return false
	}

	for _, child := range b.Children() {
		if child.Type() == "dynamic" {
			return true
		}

		if child.HasDynamicBlock() {
			return true
		}
	}

	return false
}

var (
	countRegex   = regexp.MustCompile(`\[(\d+)\]$`)
	foreachRegex = regexp.MustCompile(`\["(.*)"\]$`)
)

// Index returns the count index of the block using the name label.
// Index returns nil if the block has no count.
func (b *Block) Index() *int64 {
	m := countRegex.FindStringSubmatch(b.NameLabel())

	if len(m) > 0 {
		i, _ := strconv.ParseInt(m[1], 10, 64)

		return &i
	}

	return nil
}

// Key returns the foreach key of the block using the name label.
// Key returns nil if the block has no each key.
func (b *Block) Key() *string {
	m := foreachRegex.FindStringSubmatch(b.NameLabel())

	if len(m) > 0 {
		return &m[1]
	}

	return nil
}

func (b *Block) Label() string {
	if b == nil || b.HCLBlock == nil {
		return ""
	}

	return strings.Join(b.HCLBlock.Labels, ".")
}

func loadBlocksFromFile(file file, schema *hcl.BodySchema) (hcl.Blocks, error) {
	if schema == nil {
		schema = terraformSchemaV012
	}

	contents, _, diags := file.hclFile.Body.PartialContent(schema)
	if diags != nil && diags.HasErrors() {
		return nil, diags
	}

	if contents == nil {
		return nil, errorNoHCLContents
	}

	return contents.Blocks, nil
}

func stripCount(s string) string {
	return foreachRegex.ReplaceAllString(countRegex.ReplaceAllString(s, ""), "")
}

// BlockValueFunc defines a type that returns a set of fake/mocked values for a given block type.
type BlockValueFunc = func(b *Block) cty.Value

var (
	defaultAWSRegion = "us-east-1"
	defaultGCPRegion = "us-central1"

	blockValueFuncs = map[string]BlockValueFunc{
		"data.aws_availability_zones":  awsAvailabilityZonesValues,
		"data.google_compute_zones":    googleComputeZonesValues,
		"data.aws_region":              awsCurrentRegion,
		"data.aws_default_tags":        awsDefaultTagValues,
		"data.aws_subnets":             awsSubnetsValues,
		"resource.random_shuffle":      randomShuffleValues,
		"resource.time_static":         timeStaticValues,
		"resource.aws_launch_template": launchTemplateValues,
	}
)

// launchTemplateValues returns the values for the launch template but if the
// name attribute is not set it will generate a unique name based on the block
// address. This is done to ensure that we can properly reference the launch
// template, when users uses the name-prefix attribute to set a name.
func launchTemplateValues(b *Block) cty.Value {
	values := b.values()
	launchTemplateData := values.AsValueMap()
	v, ok := launchTemplateData["name"]

	if !ok || v.IsNull() {
		if launchTemplateData == nil {
			launchTemplateData = make(map[string]cty.Value)
		}

		h := sha256.New()
		h.Write([]byte(b.FullName()))
		addressSha := hex.EncodeToString(h.Sum(nil))
		launchTemplateData["name"] = cty.StringVal("hcl-" + addressSha)
	}

	return cty.ObjectVal(launchTemplateData)
}

// timeStaticValues mocks the values returned from resource.time_static which is
// a resource that returns the attributes of the provided rfc3339 time. If none
// is provided, it defaults to the current time.
//
// https://registry.terraform.io/providers/hashicorp/time/latest/docs/resources/static
func timeStaticValues(b *Block) cty.Value {
	now := time.Now()
	var inputDateStr string
	v := b.GetAttribute("rfc3339").Value()
	_ = gocty.FromCtyValue(v, &inputDateStr)
	if inputDateStr == "" {
		inputDateStr = now.Format(time.RFC3339)
	}

	inputDate, err := time.Parse(time.RFC3339, inputDateStr)
	if err != nil {
		inputDate = now
	}

	return cty.ObjectVal(map[string]cty.Value{
		"rfc3339": cty.StringVal(inputDateStr),
		"day":     cty.NumberIntVal(int64(inputDate.Day())),
		"hour":    cty.NumberIntVal(int64(inputDate.Hour())),
		"id":      cty.StringVal(inputDate.Format(time.RFC3339)),
		"minute":  cty.NumberIntVal(int64(inputDate.Minute())),
		"month":   cty.NumberIntVal(int64(inputDate.Month())),
		"second":  cty.NumberIntVal(int64(inputDate.Second())),
		"unix":    cty.NumberIntVal(inputDate.Unix()),
		"year":    cty.NumberIntVal(int64(inputDate.Year())),
	})
}

func awsCurrentRegion(b *Block) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"name": cty.StringVal(getRegionFromProvider(b, "aws")),
	})
}

func awsDefaultTagValues(b *Block) cty.Value {
	defaultTags := getFromProvider(b, "aws", "default_tags")
	if defaultTags.IsKnown() && defaultTags.CanIterateElements() {
		tags := defaultTags.AsValueSlice()
		if len(tags) > 0 {
			return cty.ObjectVal(map[string]cty.Value{
				"tags": tags[0].AsValueMap()["tags"],
			})
		}
	}

	return cty.ObjectVal(map[string]cty.Value{
		"tags": cty.ObjectVal(map[string]cty.Value{}),
	})
}

// awsSubnetsValues mocks the values returned from data.aws_subnets. This data
// source returns a list of subnet IDs.
// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/subnets
//
// We return a list of 3 fake subnet IDs. Although the actual number of subnets
// returned is based on the region and availability zones, we return a fixed
// "sensible" number.
func awsSubnetsValues(b *Block) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"ids": cty.ListVal([]cty.Value{
			cty.StringVal(fmt.Sprintf("subnet-1-%s", mock.Identifier)),
			cty.StringVal(fmt.Sprintf("subnet-2-%s", mock.Identifier)),
			cty.StringVal(fmt.Sprintf("subnet-3-%s", mock.Identifier)),
		}),
	})
}

// randomShuffleValues mocks the values returned from resource.random_shuffle
// https://github.com/hashicorp/terraform-provider-random/blob/main/docs/resources/shuffle.md.
// This resource uses the result_count attribute to return a slice of the input
// attribute, which is randomly sorted.
//
// We don't randomly sort it here as it's unnecessary for us, and a consistent output
// is probably better for debugging/tests.
func randomShuffleValues(b *Block) cty.Value {
	vals := b.values().AsValueMap()
	inputs := vals["input"]
	if !inputs.IsKnown() || !inputs.CanIterateElements() {
		return b.values()
	}

	elements := inputs.AsValueSlice()

	count, ok := vals["result_count"]
	if !ok {
		vals["result"] = inputs
		return cty.ObjectVal(vals)
	}

	var x = 1
	err := gocty.FromCtyValue(count, &x)
	if err != nil {
		b.logger.Debug().Err(err).Msgf("couldn't load result_count to int for random_shuffle")
	}

	if x > len(elements)-1 {
		vals["result"] = inputs
		return cty.ObjectVal(vals)
	}

	var filtered []cty.Value
	for i := 0; i < x; i++ {
		filtered = append(filtered, elements[i])
	}

	vals["result"] = cty.ListVal(filtered)
	return cty.ObjectVal(vals)
}

// googleComputeZonesValues returns the correct gcp zones for the data block region, data block provider region, or module region.
// Used zones can be found in zones_gcp.go and can be regenerated by running go generate on this file.
//
//go:generate go run ../../tools/describezones/main.go gcp
func googleComputeZonesValues(b *Block) cty.Value {
	region := getRegionFromProvider(b, "google")

	attr := b.GetAttribute("region")
	if attr != nil {
		var str string
		attrRegion := attr.Value()
		err := gocty.FromCtyValue(attrRegion, &str)
		if err == nil {
			region = str
		} else {
			b.logger.Debug().Err(err).Msgf("could not parse gcp compute zone region")
		}
	}

	if v, ok := gcpZones[region]; ok {
		return v
	}

	return gcpZones[defaultGCPRegion]
}

// awsAvailabilityZonesValues returns the correct aws zones for the data block provider region or module region.
// Used zones can be found in zones_aws.go and can be regenerated by running go generate on this file.
//
//go:generate go run ../../tools/describezones/main.go aws
func awsAvailabilityZonesValues(b *Block) cty.Value {
	region := getRegionFromProvider(b, "aws")
	if v, ok := awsZones[region]; ok {
		return v
	}

	return awsZones[defaultAWSRegion]
}

func getRegionFromProvider(b *Block, provider string) string {
	val := getFromProvider(b, provider, "region")

	var str string
	err := gocty.FromCtyValue(val, &str)
	if err != nil {
		if provider == "google" {
			return defaultGCPRegion
		}

		return defaultAWSRegion
	}

	return str
}

// getFromProvider returns the value of the given key from provider the block is
// associated with. This is either the default provider or the provider
// explicitly set on the block (e.g. if a resource has a "provider" attribute).
func getFromProvider(b *Block, provider, key string) cty.Value {
	attr := b.GetAttribute("provider")
	if attr != nil {
		v := attr.Value()
		if v.Type().IsObjectType() {
			m := v.AsValueMap()
			if r, ok := m[key]; ok {
				return r
			}
		}
	}

	providerVal := b.Context().Get(provider)
	if providerVal.Type().IsObjectType() {
		m := providerVal.AsValueMap()
		if r, ok := m[key]; ok {
			return r
		}
	}

	return cty.NilVal
}

func usesProviderConfiguration(b *Block) bool {
	return b.Type() == "resource" || b.Type() == "data"
}

func supportsDependsOn(b *Block) bool {
	return b.Type() == "resource" || b.Type() == "data" || b.Type() == "module"
}

func supportsLifecycle(b *Block) bool {
	return b.Type() == "resource" || b.Type() == "data"
}
