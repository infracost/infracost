package hcl

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VertexData struct {
	logger    *logrus.Entry
	block     *Block
	evaluator *Evaluator
}

func (v *VertexData) ID() string {
	return v.block.FullName()
}

func (v *VertexData) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexData) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexData) References() []VertexReference {
	return referencesForBlock(v.block)
}

func (v *VertexData) Evaluate() error {
	if len(v.block.Labels()) < 2 {
		return fmt.Errorf("data block %s has no label", v.ID())
	}

	var existingVals map[string]cty.Value
	existingCtx := v.evaluator.ctx.Get(v.block.TypeLabel())
	if !existingCtx.IsNull() {
		existingVals = existingCtx.AsValueMap()
	} else {
		existingVals = make(map[string]cty.Value)
	}

	val := v.evaluator.evaluateResource(v.block, existingVals)

	v.logger.Debugf("adding data %s to the evaluation context", v.ID())
	v.evaluator.ctx.SetByDot(val, fmt.Sprintf("data.%s", v.block.TypeLabel()))

	return nil
}

func (v *VertexData) Expand() ([]*Block, error) {
	visitMu.Lock()
	defer visitMu.Unlock()

	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockCounts(expanded)
	expanded = v.evaluator.expandBlockForEaches(expanded)
	expanded = v.evaluator.expandDynamicBlocks(expanded...)

	v.block.expanded = true

	return expanded, nil
}
