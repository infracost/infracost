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

func (v *VertexData) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexData) References() []string {
	return referencesForBlock(v.block)
}

func (v *VertexData) Evaluate() error {
	if len(v.block.Labels()) < 2 {
		return fmt.Errorf("data block %s has no label", v.ID())
	}

	val := v.evaluator.evaluateResource(v.block, map[string]cty.Value{})

	v.logger.Debugf("adding data %s to the evaluation context", v.ID())
	key := fmt.Sprintf("data.%s", v.block.LocalName())
	v.evaluator.ctx.Set(val, key)

	return nil
}

func (v *VertexData) Expand() ([]*Block, error) {
	visitMu.Lock()
	defer visitMu.Unlock()

	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockCounts(expanded)
	expanded = v.evaluator.expandBlockForEaches(expanded)

	return expanded, nil
}
