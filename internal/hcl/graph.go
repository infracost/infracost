package hcl

import (
	"fmt"
	"strings"

	"github.com/heimdalr/dag"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type Graph struct {
	dag       *dag.DAG
	evaluator *Evaluator
	logger    *logrus.Entry
}

func NewGraph(evaluator *Evaluator, logger *logrus.Entry) *Graph {
	return &Graph{
		dag:       dag.NewDAG(),
		evaluator: evaluator,
		logger:    logger,
	}
}

func (g *Graph) Populate() error {
	for _, block := range g.evaluator.module.Blocks {
		switch block.Type() {
		case "locals":
			for _, attr := range block.GetAttributes() {
				err := g.dag.AddVertexByID(fmt.Sprintf("locals.%s", attr.Name()), attr)
				if err != nil {
					return err
				}
			}
		default:
			err := g.dag.AddVertexByID(block.FullName(), block)
			if err != nil {
				return err
			}
		}
	}

	for id, vertex := range g.dag.GetVertices() {
		switch val := vertex.(type) {
		case *Attribute:
			attr := val

			for _, ref := range attr.AllReferences() {
				err := g.dag.AddEdge(ref.String(), id)
				if err != nil {
					g.logger.Errorf("error adding edge: %s", err)
				}
			}
		case *Block:
			block := val

			for _, attr := range block.GetAttributes() {
				for _, ref := range attr.AllReferences() {
					err := g.dag.AddEdge(ref.String(), id)
					if err != nil {
						g.logger.Errorf("error adding edge: %s", err)
					}
				}
			}

			for _, childBlock := range block.Children() {
				for _, attr := range childBlock.GetAttributes() {
					for _, ref := range attr.AllReferences() {
						err := g.dag.AddEdge(ref.String(), id)
						if err != nil {
							g.logger.Errorf("error adding edge: %s", err)
						}
					}
				}
			}
		}
	}

	j, _ := g.asJSON()
	g.logger.Debugf("graph: %s", j)

	return nil
}

func (g *Graph) asJSON() ([]byte, error) {
	return g.dag.MarshalJSON()
}

func (g *Graph) Walk() {
	v := NewGraphVisitor(g.evaluator, g.logger)
	g.dag.DFSWalk(v)
	g.evaluator.module.Blocks = v.filteredBlocks
}

type GraphVisitor struct {
	evaluator      *Evaluator
	logger         *logrus.Entry
	filteredBlocks []*Block
}

func NewGraphVisitor(evaluator *Evaluator, logger *logrus.Entry) *GraphVisitor {
	return &GraphVisitor{
		evaluator: evaluator,
		logger:    logger,
	}
}

func (v *GraphVisitor) Visit(vertexer dag.Vertexer) {
	id, vertex := vertexer.Vertex()

	v.logger.Debugf("visiting %s", id)

	ctx := v.evaluator.ctx

	// Check if this is already in context
	ctxVal := ctx.Get(id)
	if !ctxVal.IsNull() {
		v.logger.Debugf("%s already in context", id)
		return
	}

	switch val := vertex.(type) {
	case *Attribute:
		v.visitAttribute(id, val)
	case *Block:
		v.visitBlock(id, val)
	}
}

func (v *GraphVisitor) visitAttribute(id string, attr *Attribute) {
	ctx := v.evaluator.ctx

	v.logger.Debugf("adding attribute %s to the evaluation context", id)

	parts := strings.Split(id, ".")
	if len(parts) > 1 && parts[0] == "locals" {
		key := fmt.Sprintf("local.%s", strings.Join(parts[1:], "."))
		ctx.SetByDot(attr.Value(), key)
	}
}

func (v *GraphVisitor) visitBlock(id string, b *Block) {
	ctx := v.evaluator.ctx

	switch b.Type() {
	case "variable": // variables are special in that their value comes from the "default" attribute
		val, err := v.evaluator.evaluateVariable(b)
		if err != nil {
			v.logger.WithError(err).Debugf("could not evaluate variable %s ignoring", b.FullName())
			return
		}

		v.logger.Debugf("adding variable %s to the evaluation context", id)
		key := fmt.Sprintf("var.%s", b.Label())
		ctx.SetByDot(val, key)
	case "output":
		val, err := v.evaluator.evaluateOutput(b)
		if err != nil {
			v.logger.WithError(err).Debugf("could not evaluate output %s ignoring", b.FullName())
			return
		}

		v.logger.Debugf("adding output %s to the evaluation context", id)
		key := fmt.Sprintf("output.%s", b.Label())
		ctx.SetByDot(val, key)
	case "provider":
		provider := b.Label()
		if provider == "" {
			return
		}

		ctx.Set(v.evaluator.evaluateProvider(b, map[string]cty.Value{}), id)
	case "module":
		if b.Label() == "" {
			return
		}

		v.logger.Debugf("adding module %s to the evaluation context", b.Label())
		ctx.Set(b.Values(), id)
	case "resource":
		if len(b.Labels()) < 2 {
			return
		}

		val := v.evaluator.evaluateResource(b, map[string]cty.Value{})

		v.logger.Debugf("adding resource %s to the evaluation context", id)
		ctx.SetByDot(val, id)
	case "data":
		if len(b.Labels()) < 2 {
			return
		}

		val := v.evaluator.evaluateResource(b, map[string]cty.Value{})

		v.logger.Debugf("adding data %s to the evaluation context", id)
		key := fmt.Sprintf("data.%s", b.Label())
		ctx.Set(val, key)
	}

	v.logger.Debugf("adding %s to the filtered blocks", id)

	expanded := []*Block{b}
	expanded = v.evaluator.expandBlockForEaches(expanded)
	expanded = v.evaluator.expandBlockCounts(expanded)
	v.filteredBlocks = append(v.filteredBlocks, expanded...)
}
