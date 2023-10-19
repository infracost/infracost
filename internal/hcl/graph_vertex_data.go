package hcl

import (
	"fmt"
	"sync"

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

func (v *VertexData) References() []VertexReference {
	return referencesForBlock(v.block)
}

func (v *VertexData) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	err := v.evaluate()
	if err != nil {
		return fmt.Errorf("could not evaluate data block %q", v.ID())
	}

	expanded, err := v.expand()
	if err != nil {
		return fmt.Errorf("could not expand data block %q", v.ID())
	}

	v.evaluator.AddFilteredBlocks(expanded...)
	return nil
}

func (v *VertexData) evaluate() error {
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

func (v *VertexData) expand() ([]*Block, error) {
	expanded := []*Block{v.block}
	expanded = v.evaluator.expandBlockCounts(expanded)
	expanded = v.evaluator.expandBlockForEaches(expanded)
	expanded = v.evaluator.expandDynamicBlocks(expanded...)

	v.block.expanded = true

	return expanded, nil
}
