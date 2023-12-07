package hcl

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/schema"
)

var (
	addrSplitModuleRegex = regexp.MustCompile(`^((?:module\.[^.]+\.?)+)\.(.*)$`)
)

type ModuleConfig struct {
	name            string
	moduleCall      *ModuleCall
	evaluator       *Evaluator
	parentEvaluator *Evaluator
}

type ModuleConfigs struct {
	configs map[string][]ModuleConfig
	mu      sync.RWMutex
}

func NewModuleConfigs() *ModuleConfigs {
	return &ModuleConfigs{
		configs: make(map[string][]ModuleConfig),
		mu:      sync.RWMutex{},
	}
}

func (m *ModuleConfigs) Add(moduleAddress string, moduleConfig ModuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.configs[moduleAddress]; !ok {
		m.configs[moduleAddress] = []ModuleConfig{}
	}

	m.configs[moduleAddress] = append(m.configs[moduleAddress], moduleConfig)
}

func (m *ModuleConfigs) Get(moduleAddress string) []ModuleConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.configs[moduleAddress]
}

type Graph struct {
	dag           *dag.DAG
	logger        zerolog.Logger
	rootVertex    Vertex
	vertexMutex   *sync.Mutex
	moduleConfigs *ModuleConfigs
}

// Vertex interface represents a graph vertex with unique identifier and module
// address. It also provides methods for visiting the vertex, retrieving
// references, and getting the vertex ID.
type Vertex interface {
	ID() string
	ModuleAddress() string
	Visit(mutex *sync.Mutex) (interface{}, error)
	References() []VertexReference
}

// VertexReference represents a reference to a vertex in a graph. It contains
// information about the module address, attribute name, type, and key of the
// referenced vertex.
type VertexReference struct {
	ModuleAddress string
	AttributeName string
	Type          string
	Key           string
}

func NewGraphWithRoot(logger zerolog.Logger) (*Graph, error) {
	g := &Graph{
		dag:           dag.NewDAG(),
		logger:        logger,
		moduleConfigs: NewModuleConfigs(),
		vertexMutex:   &sync.Mutex{},
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

	blocks, err := g.loadAllBlocks(evaluator)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		switch block.Type() {
		case "locals":
			for _, attr := range block.GetAttributes() {
				vertexes = append(vertexes, &VertexLocal{
					logger:        g.logger,
					moduleConfigs: g.moduleConfigs,
					block:         block,
					attr:          attr,
				})
			}
		case "module":
			vertexes = append(vertexes, &VertexModuleCall{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
			vertexes = append(vertexes, &VertexModuleExit{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
		case "variable":
			vertexes = append(vertexes, &VertexVariable{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
		case "output":
			vertexes = append(vertexes, &VertexOutput{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
		case "provider":
			vertexes = append(vertexes, &VertexProvider{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
		case "resource":
			vertexes = append(vertexes, &VertexResource{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
		case "data":
			vertexes = append(vertexes, &VertexData{
				logger:        g.logger,
				moduleConfigs: g.moduleConfigs,
				block:         block,
			})
		}
	}

	for _, vertex := range vertexes {
		id := vertex.ID()

		g.logger.Debug().Msgf("adding vertex: %s", id)
		err := g.dag.AddVertexByID(id, vertex)
		if err != nil {
			// We don't actually mind if blocks are added multiple times
			// since this helps us support cases like _override.tf files
			// and in-progress changes.
			g.logger.Debug().Err(err).Msgf("error adding vertex %q", id)
		}
	}

	edges := make([]dag.EdgeInput, 0)

	for _, vertex := range vertexes {
		id := vertex.ID()
		modAddr := vertex.ModuleAddress()

		if modAddr == "" {
			g.logger.Debug().Msgf("adding edge: %s, %s", g.rootVertex.ID(), id)
			edges = append(edges, dag.EdgeInput{
				SrcID: g.rootVertex.ID(),
				DstID: id,
			})
		} else {
			// Add the module call edge
			g.logger.Debug().Msgf("adding edge: %s, %s", moduleCallID(modAddr), id)
			edges = append(edges, dag.EdgeInput{
				SrcID: moduleCallID(modAddr),
				DstID: id,
			})

			// Add the module exit edge
			g.logger.Debug().Msgf("adding edge: %s, %s", id, modAddr)
			edges = append(edges, dag.EdgeInput{
				SrcID: id,
				DstID: modAddr,
			})
		}

		for _, ref := range vertex.References() {
			var srcID string

			parts := strings.Split(ref.Key, ".")
			idx := len(parts)

			// data references should always have a length of 3
			// provider references might have a length of 3 (if using an alias) or 2 (if not).
			if (strings.HasPrefix(ref.Key, "data.") || strings.HasPrefix(ref.Key, "provider.")) && len(parts) >= 3 {
				// Source ID is the first 3 parts or less if the length of parts is less than 3
				idx = 3
			} else if len(parts) >= 2 {
				// variable, local, resources and output references should all have length 2
				idx = 2
			}
			srcID = strings.Join(parts[:idx], ".")

			// Don't add the module prefix for providers since they are
			// evaluated in the root module
			if !strings.HasPrefix(srcID, "provider.") && ref.ModuleAddress != "" {
				modAddress := stripCount(ref.ModuleAddress)

				srcID = fmt.Sprintf("%s.%s", modAddress, srcID)
			}

			dstID := id

			// If the vertex is a module call and the attribute for the reference isn't a module call
			// arg (source, count, for_each), etc, then the reference is a module input and points to
			// a variable block within the module. In that case we should add the edge directly to that
			// variable block instead of the module call. We need to do this to avoid a circular dependency
			// where a module input can depend on a module output, e.g:
			//
			// module "my_module" {
			//   source = "./foo"
			//   foo = module.my_module.bar
			// }
			if _, ok := vertex.(*VertexModuleCall); ok {
				if ref.AttributeName != "" && attrIsVarInput(ref.AttributeName) {
					dstID = fmt.Sprintf("%s.variable.%s", stripModuleCallPrefix(id), ref.AttributeName)

					// Check this vertex exists
					_, err := g.dag.GetVertex(dstID)
					if err != nil {
						g.logger.Debug().Err(err).Msgf("ignoring edge %s, %s because the destination vertex doesn't exist", srcID, dstID)
						continue
					}
				}
			}

			// Strip the count/index suffix from the source ID
			srcID = stripCount(srcID)

			if srcID == dstID {
				continue
			}

			// Check if the source vertex exists
			_, err := g.dag.GetVertex(srcID)
			if err == nil {
				g.logger.Debug().Msgf("adding edge: %s, %s", srcID, dstID)
				edges = append(edges, dag.EdgeInput{
					SrcID: srcID,
					DstID: dstID,
				})

				continue
			}

			// If the source vertex doesn't exist, it might be a module output attribute,
			// so we need to check if the module output exists and add an edge from that
			// to the current vertex instead.
			if ref.ModuleAddress != "" && stripCount(ref.ModuleAddress) != modAddr {
				srcID = fmt.Sprintf("%s.%s", stripCount(ref.ModuleAddress), parts[0])

				// Check if the source vertex exists
				_, err := g.dag.GetVertex(srcID)
				if err == nil {
					g.logger.Debug().Msgf("adding edge: %s, %s", srcID, dstID)
					edges = append(edges, dag.EdgeInput{
						SrcID: srcID,
						DstID: dstID,
					})

					continue
				}
			}
		}
	}

	err = g.dag.AddEdges(edges)
	if err != nil {
		return fmt.Errorf("error adding edges %w", err)
	}

	// Setup initial context
	g.moduleConfigs.Add("", ModuleConfig{
		name:       "",
		moduleCall: nil,
		evaluator:  evaluator,
	})

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

// MarshalJSON returns the JSON encoding of the graph.
func (g *Graph) MarshalJSON() ([]byte, error) {
	return g.dag.MarshalJSON()
}

// Walk traverses the graph and collects the resources into a list of
// ResourceData. It uses the DescendantsFlow method of the underlying graph to
// visit each vertex in the graph.
//
// The first DescendantsFlow call executes the NewGraphEvaluateVisitFunc callback
// function, which calls the Visit method of each vertex in the graph, evaluating
// the resources so that they have the correct values and expansion.
//
// The second DescendantsFlow call executes the Visit method of
// ResourceCollectorVisitor, which transforms each vertex of type *VertexResource
// into a slice of schema.Resource. After all the resources have been collected,
// we then populate the references, between each resource by checking the
// references for each returned Resource.
func (g *Graph) Walk() []ResourceData {
	_, _ = g.dag.DescendantsFlow(g.rootVertex.ID(), nil, NewGraphEvaluateVisitFunc(g.logger, g.vertexMutex))

	coll := &ResourceCollectorVisitor{mu: &sync.Mutex{}}
	_, _ = g.dag.DescendantsFlow(g.rootVertex.ID(), nil, coll.Visit)

	coll.BuildResourceReferences()
	return coll.resources
}

// Run executes the graph evaluation process with the given evaluator and returns
// the computed ResourceData and any error encountered during the process. It
// populates the graph using the provided evaluator, reduces transitively, and
// performs a graph walk to collect the graph into a list of ResourceData.
func (g *Graph) Run(evaluator *Evaluator) ([]ResourceData, error) {
	err := g.Populate(evaluator)
	if err != nil {
		return nil, err
	}

	g.ReduceTransitively()
	return g.Walk(), nil
}

// NewGraphEvaluateVisitFunc returns a callback function that can be used in the
// DescendantsFlow method of DAG to visit the graph vertices and evaluate them.
//
// The callback function defers Visit functionality to each Vertex. The
// vertexMutex is used to synchronize access to the vertices. This is necessary
// as the global Context will panic with concurrent map/read write if we don't
// lock node visits.
func NewGraphEvaluateVisitFunc(logger zerolog.Logger, vertexMutex *sync.Mutex) dag.FlowCallback {
	return func(d *dag.DAG, id string, parentResults []dag.FlowResult) (interface{}, error) {
		logger.Debug().Msgf("visiting vertex %q", id)

		vertex, _ := d.GetVertex(id)
		vert := vertex.(Vertex)

		return vert.Visit(vertexMutex)
	}
}

// ResourceCollectorVisitor implements a Visitor for the DAG DescendantsFlow. It
// is designed to transform the DAG into a simple slice of resources.
type ResourceCollectorVisitor struct {
	resources []ResourceData
	mu        *sync.Mutex
}

// Visit implements the DescendantsFlow callback method of DAG. Visit traverses
// the DAG and collects all the vertices into a slice ResourceData which
// represent all the applicable terraform resources for the DAG.
//
// Visit should be used after the DAG has been evaluated using a
// GraphEvaluateVisitFunc.
func (v *ResourceCollectorVisitor) Visit(d *dag.DAG, id string, parentResults []dag.FlowResult) (interface{}, error) {
	vertex, _ := d.GetVertex(id)
	vert, ok := vertex.(*VertexResource)
	if !ok {
		return nil, nil
	}

	resources := vert.TransformToSchemaResources()
	v.mu.Lock()
	v.resources = append(v.resources, resources...)
	v.mu.Unlock()

	return resources, nil
}

// BuildResourceReferences iterates over the collected resources and builds the
// references between them. It creates a map (resMap) to store the address of
// each resource. Then, for each resource, it retrieves the vertices referenced
// by the resource's block. If the referenced vertex is of type "resource", it
// performs the following steps:
//
//   - Fetches the stored references for the
//     attribute.
//   - Retrieves the referenced resource from resMap. - Appends the
//     referenced resource to the stored references.
//   - Adds a reverse reference from
//     the referenced resource to the current resource.
func (v *ResourceCollectorVisitor) BuildResourceReferences() {
	resMap := make(map[string]*schema.ResourceData)

	for _, resource := range v.resources {
		resMap[resource.Data.Address] = resource.Data
	}

	for _, resource := range v.resources {
		refs := resource.Block.VerticesReferenced()
		for _, ref := range refs {
			if ref.Type != "resource" {
				continue
			}

			// @TODO handle # and * symbols for referenes
			storedRefs := resource.Data.ReferencesMap[ref.AttributeName]
			referenced := resMap[ref.Key]
			if referenced == nil {
				continue
			}

			// alter the resource references to contain reverse references to the resource
			// that was used in the original reference.
			resource.Data.ReferencesMap[ref.AttributeName] = append(storedRefs, referenced)
			reverseRefKey := resource.Data.Type + "." + ref.AttributeName
			referenced.ReferencesMap[reverseRefKey] = append(referenced.ReferencesMap[reverseRefKey], resource.Data)
		}
	}
}

func (g *Graph) loadAllBlocks(evaluator *Evaluator) ([]*Block, error) {
	return g.loadBlocksForModule(evaluator)
}

func (g *Graph) loadBlocksForModule(evaluator *Evaluator) ([]*Block, error) {
	var blocks []*Block

	for _, block := range evaluator.module.Blocks {
		blocks = append(blocks, block)

		if block.Type() == "module" {
			modCall, err := evaluator.loadModule(block)
			if err != nil {
				return nil, fmt.Errorf("could not load module %q", block.FullName())
			}

			moduleEvaluator := NewEvaluator(
				*modCall.Module,
				evaluator.workingDir,
				map[string]cty.Value{},
				evaluator.moduleMetadata,
				map[string]map[string]cty.Value{},
				evaluator.workspace,
				evaluator.blockBuilder,
				nil,
				evaluator.logger,
				&Context{},
			)

			modBlocks, err := g.loadBlocksForModule(moduleEvaluator)
			if err != nil {
				return nil, fmt.Errorf("could not load blocks for module %q", block.FullName())
			}

			blocks = append(blocks, modBlocks...)
		}
	}

	return blocks, nil
}
