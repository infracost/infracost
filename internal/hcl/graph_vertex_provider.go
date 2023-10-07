package hcl

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VertexProvider struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
}

func (v *VertexProvider) ID() string {
	alias := v.block.GetAttribute("alias")
	id := v.block.FullName()
	if alias != nil {
		id = fmt.Sprintf("%s.%s", v.block.FullName(), alias.Value().AsString())
	}

	return id
}

func (v *VertexProvider) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexProvider) References() []string {
	return referencesForBlock(v.block)
}

func (v *VertexProvider) Evaluate() error {
	provider := v.block.Label()
	if provider == "" {
		return fmt.Errorf("provider block %s has no label", v.ID())
	}

	var existingVals map[string]cty.Value
	existingCtx := v.evaluator.ctx.Get(v.block.TypeLabel())
	if !existingCtx.IsNull() {
		existingVals = existingCtx.AsValueMap()
	} else {
		existingVals = make(map[string]cty.Value)
	}

	val := v.evaluator.evaluateProvider(v.block, existingVals)

	v.logger.Debugf("adding %s to the evaluation context", v.ID())
	v.evaluator.ctx.Set(val, v.block.Label())

	return nil
}

func (v *VertexProvider) Expand() ([]*Block, error) {
	return []*Block{v.block}, nil
}
