package hcl

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VertexResource struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
}

func (v *VertexResource) ID() string {
	return v.block.FullName()
}

func (v *VertexResource) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexResource) References() []string {
	return referencesForBlock(v.block)
}

func (v *VertexResource) Evaluate() error {
	if len(v.block.Labels()) < 2 {
		return fmt.Errorf("resource block %s has no label", v.ID())
	}

	var existingVals map[string]cty.Value
	existingCtx := v.evaluator.ctx.Get(v.block.TypeLabel())
	if !existingCtx.IsNull() {
		existingVals = existingCtx.AsValueMap()
	} else {
		existingVals = make(map[string]cty.Value)
	}

	val := v.evaluator.evaluateResource(v.block, existingVals)

	v.logger.Debugf("adding resource %s to the evaluation context", v.ID())
	v.evaluator.ctx.SetByDot(val, v.block.TypeLabel())

	return nil
}

func (v *VertexResource) Expand() ([]*Block, error) {
	visitMu.Lock()
	defer func() {
		visitMu.Unlock()
	}()

	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockCounts(expanded)
	expanded = v.evaluator.expandBlockForEaches(expanded)

	return expanded, nil
}
