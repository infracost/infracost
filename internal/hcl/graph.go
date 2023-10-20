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
	addrSplitModuleRegex = regexp.MustCompile(`^((?:module\.[^.]+\.?)+)\.(.*)$`)
)

type Graph struct {
	dag         *dag.DAG
	logger      *logrus.Entry
	rootVertex  Vertex
	vertexMutex *sync.Mutex
}

type Vertex interface {
	ID() string
	ModuleAddress() string
	Visit(mutex *sync.Mutex) error
	References() []VertexReference
}

type VertexReference struct {
	ModuleAddress string
	Key           string
}

func NewGraphWithRoot(logger *logrus.Entry, vertexMutex *sync.Mutex) (*Graph, error) {
	if vertexMutex == nil {
		vertexMutex = &sync.Mutex{}
	}
	g := &Graph{
		dag:         dag.NewDAG(),
		logger:      logger,
		vertexMutex: vertexMutex,
	}

	g.rootVertex = &VertexRoot{}

	err := g.dag.AddVertexByID(g.rootVertex.ID(), g.rootVertex)
	if err != nil {
		return nil, fmt.Errorf("error adding vertex %q %w", g.rootVertex.ID(), err)
	}

	return g, nil
}

func (g *Graph) ReduceTransitively() {
	g.dag.ReduceTransitively()
}

func (g *Graph) Populate(evaluator *Evaluator) error {
	var vertexes []Vertex

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
			return fmt.Errorf("error adding vertex %q %w", vertex.ID(), err)
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
	v := NewGraphVisitor(g.logger, g.vertexMutex)

	flowCallback := func(d *dag.DAG, id string, parentResults []dag.FlowResult) (interface{}, error) {
		vertex, _ := d.GetVertex(id)

		v.Visit(id, vertex)

		return vertex, nil
	}

	_, _ = g.dag.DescendantsFlow(g.rootVertex.ID(), nil, flowCallback)
}

func (g *Graph) Run(evaluator *Evaluator) (*Module, error) {
	err := g.Populate(evaluator)
	if err != nil {
		return nil, err
	}

	g.ReduceTransitively()
	g.Walk()
	evaluator.module.Blocks = evaluator.filteredBlocks
	evaluator.module = *evaluator.collectModules()

	return &evaluator.module, nil
}

type GraphVisitor struct {
	logger      *logrus.Entry
	vertexMutex *sync.Mutex
}

func NewGraphVisitor(logger *logrus.Entry, vertexMutex *sync.Mutex) *GraphVisitor {
	return &GraphVisitor{
		logger:      logger,
		vertexMutex: vertexMutex,
	}
}

func (v *GraphVisitor) Visit(id string, vertex interface{}) {
	v.logger.Debugf("visiting vertex %q", id)

	vert := vertex.(Vertex)
	err := vert.Visit(v.vertexMutex)
	if err != nil {
		v.logger.WithError(err).Debugf("ignoring vertex %q because an error was encountered", id)
	}
}
