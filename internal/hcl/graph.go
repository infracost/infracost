package hcl

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

var (
	visitMu              sync.Mutex
	addrSplitModuleRegex = regexp.MustCompile(`^((?:module\.[^.]+\.?)+)\.(.*)$`)
)

type Graph struct {
	dag        *dag.DAG
	logger     *logrus.Entry
	rootVertex Vertex
}

type Vertex interface {
	ID() string
	ModuleAddress() string
	Evaluator() *Evaluator
	References() []VertexReference
	Evaluate() error
	Expand() ([]*Block, error)
}

type VertexReference struct {
	ModuleAddress string
	Key           string
}

func NewGraph(logger *logrus.Entry) *Graph {
	return &Graph{
		dag:    dag.NewDAG(),
		logger: logger,
	}
}

func NewGraphWithRoot(logger *logrus.Entry) (*Graph, error) {
	g := NewGraph(logger)

	g.rootVertex = &VertexRoot{}

	err := g.dag.AddVertexByID(g.rootVertex.ID(), g.rootVertex)
	if err != nil {
		return nil, fmt.Errorf("error adding vertex: %w", err)
	}

	return g, nil
}

func (g *Graph) ReduceTransitively() {
	g.dag.ReduceTransitively()
}

func (g *Graph) Populate(evaluator *Evaluator) error {
	vertexes := []Vertex{}

	for _, block := range evaluator.module.Blocks {
		switch block.Type() {
		case "locals":
			for _, attr := range block.GetAttributes() {
				vertexes = append(vertexes, &VertexLocal{
					logger:    g.logger,
					evaluator: evaluator,
					block:     block,
					attr:      attr,
				})
			}
		case "module":
			vertexes = append(vertexes, &VertexModule{
				logger:    g.logger,
				evaluator: evaluator,
				block:     block,
			})
		case "variable":
			vertexes = append(vertexes, &VertexVariable{
				logger:    g.logger,
				evaluator: evaluator,
				block:     block,
			})
		case "output":
			vertexes = append(vertexes, &VertexOutput{
				logger:    g.logger,
				evaluator: evaluator,
				block:     block,
			})
		case "provider":
			vertexes = append(vertexes, &VertexProvider{
				logger:    g.logger,
				evaluator: evaluator,
				block:     block,
			})
		case "resource":
			vertexes = append(vertexes, &VertexResource{
				logger:    g.logger,
				evaluator: evaluator,
				block:     block,
			})
		case "data":
			vertexes = append(vertexes, &VertexData{
				logger:    g.logger,
				evaluator: evaluator,
				block:     block,
			})
		}
	}

	for _, vertex := range vertexes {
		err := g.dag.AddVertexByID(vertex.ID(), vertex)
		if err != nil {
			return err
		}
	}

	for _, vertex := range vertexes {
		g.logger.Debugf("adding edge: %s, %s", g.rootVertex.ID(), vertex.ID())

		err := g.dag.AddEdge(g.rootVertex.ID(), vertex.ID())
		if err != nil {
			g.logger.Debugf("error adding edge: %s", err)
		}

		for _, ref := range vertex.References() {
			var srcId string

			parts := strings.Split(ref.Key, ".")
			idx := len(parts)

			// If the reference is to an attribute of a block then we want to
			// use the block part as the source ID, except for locals where we
			// want to use the first-level attribute name
			if strings.HasPrefix(ref.Key, "data.") && len(parts) >= 3 {
				// Source ID is the first 3 parts or less if the length of parts is less than 3
				idx = 3
			} else if len(parts) >= 2 {
				// resources and outputs don't have a type prefix in the references
				idx = 2
			}
			srcId = strings.Join(parts[:idx], ".")

			// Don't add the module prefix for providers since they are
			// evaluated in the root module
			if !strings.HasPrefix(srcId, "provider.") && ref.ModuleAddress != "" {
				srcId = fmt.Sprintf("%s.%s", ref.ModuleAddress, srcId)
			}

			// If the reference points to a different module then we want to add
			// a dependency for that module call
			if ref.ModuleAddress != vertex.ModuleAddress() {
				srcId = ref.ModuleAddress
			}

			// Strip the count/index suffix from the source ID
			srcId = stripCount(srcId)

			if srcId == vertex.ID() {
				continue
			}

			g.logger.Debugf("adding edge: %s, %s", srcId, vertex.ID())
			err := g.dag.AddEdge(srcId, vertex.ID())
			if err != nil {
				g.logger.Debugf("error adding edge: %s", err)
			}
		}
	}

	// Setup initial context
	evaluator.ctx.Set(cty.ObjectVal(map[string]cty.Value{}), "var")
	evaluator.ctx.Set(cty.ObjectVal(map[string]cty.Value{}), "data")
	evaluator.ctx.Set(cty.ObjectVal(map[string]cty.Value{}), "local")
	evaluator.ctx.Set(cty.ObjectVal(map[string]cty.Value{}), "output")

	// TODO: is there a better way of doing this?
	// Add the locals block to the list of filtered blocks. These won't get
	// added when we walk the graph since locals are evaluated per attribute
	// and not per block, so we need to make sure this is done here.
	evaluator.AddFilteredBlocks(evaluator.module.Blocks.OfType("locals")...)

	return nil
}

func (g *Graph) AsJSON() ([]byte, error) {
	return g.dag.MarshalJSON()
}

func (g *Graph) Walk() {
	v := NewGraphVisitor(g.logger)

	flowCallback := func(d *dag.DAG, id string, parentResults []dag.FlowResult) (interface{}, error) {
		vertex, _ := d.GetVertex(id)

		v.Visit(id, vertex)

		return vertex, nil
	}

	_, _ = g.dag.DescendantsFlow(g.rootVertex.ID(), nil, flowCallback)
}

type GraphVisitor struct {
	logger *logrus.Entry
}

func NewGraphVisitor(logger *logrus.Entry) *GraphVisitor {
	return &GraphVisitor{
		logger: logger,
	}
}

func (v *GraphVisitor) Visit(id string, vertex interface{}) {
	v.logger.Debugf("visiting %s", id)

	vert := vertex.(Vertex)

	visitMu.Lock()
	err := vert.Evaluate()
	visitMu.Unlock()
	if err != nil {
		v.logger.WithError(err).Debugf("could not evaluate %s ignoring", id)
		return
	}

	blocks, err := vert.Expand()
	if err != nil {
		v.logger.WithError(err).Debugf("could not expand %s ignoring", id)
		return
	}

	ve := vert.Evaluator()
	if ve != nil && vert.Evaluator() != nil {
		vert.Evaluator().AddFilteredBlocks(blocks...)
	}
}

func referencesForBlock(b *Block) []VertexReference {
	refs := []VertexReference{}

	hasProviderAttr := false

	for _, attr := range b.GetAttributes() {
		if attr.Name() == "provider" {
			hasProviderAttr = true
		}

		refs = append(refs, referencesForAttribute(b, attr)...)
	}

	for _, childBlock := range b.Children() {
		refs = append(refs, referencesForBlock(childBlock)...)
	}

	if !hasProviderAttr && (b.Type() == "resource" || b.Type() == "data") {
		providerName := b.Provider()
		if providerName != "" {
			refs = append(refs, VertexReference{
				Key: fmt.Sprintf("provider.%s", providerName),
			})
		}
	}

	return refs
}

func referencesForAttribute(b *Block, a *Attribute) []VertexReference {
	refs := []VertexReference{}

	for _, ref := range a.AllReferences() {
		key := ref.String()

		if shouldSkipRef(b, a, key) {
			continue
		}

		if (b.Type() == "resource" || b.Type() == "data") && a.Name() == "provider" {
			key = fmt.Sprintf("provider.%s", key)
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
			Key:           otherPart,
		})
	}

	return refs
}

func shouldSkipRef(block *Block, attr *Attribute, key string) bool {
	if key == "count.index" || key == "each.key" || key == "each.value" || strings.HasSuffix(key, ".") {
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
