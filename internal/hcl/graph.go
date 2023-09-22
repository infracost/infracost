package hcl

import (
	"fmt"
	"sync"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
)

var visitMu sync.Mutex

type Graph struct {
	dag        *dag.DAG
	logger     *logrus.Entry
	rootVertex Vertex
}

type Vertex interface {
	ID() string
	Evaluator() *Evaluator
	References() []string
	Evaluate() error
	Expand() ([]*Block, error)
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
		err := g.dag.AddEdge(g.rootVertex.ID(), vertex.ID())
		if err != nil {
			g.logger.Errorf("error adding edge: %s", err)
		}

		refIds := vertex.References()
		for _, refId := range refIds {
			err := g.dag.AddEdge(refId, vertex.ID())
			if err != nil {
				g.logger.Errorf("error adding edge: %s", err)
			}
		}
	}

	return nil
}

func (g *Graph) asJSON() ([]byte, error) {
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
	if ve != nil {
		vert.Evaluator().AddFilteredBlocks(blocks...)
	}
}

func referencesForBlock(b *Block) []string {
	refIds := []string{}

	for _, attr := range b.GetAttributes() {
		for _, ref := range attr.AllReferences() {
			refId := ref.String()
			if b.ModuleName() != "" {
				refId = fmt.Sprintf("%s.%s", b.ModuleName(), refId)
			}
			refIds = append(refIds, refId)
		}
	}

	for _, childBlock := range b.Children() {
		for _, attr := range childBlock.GetAttributes() {
			for _, ref := range attr.AllReferences() {
				refId := ref.String()
				if childBlock.ModuleName() != "" {
					refId = fmt.Sprintf("%s.%s", childBlock.ModuleName(), refId)
				}
				refIds = append(refIds, refId)
			}
		}
	}

	return refIds
}

func referencesForAttribute(a *Attribute) []string {
	refIds := []string{}

	for _, ref := range a.AllReferences() {
		refIds = append(refIds, ref.String())
	}

	return refIds
}
