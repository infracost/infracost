package hcl

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type VertexVariable struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
}

func (v *VertexVariable) ID() string {
	return v.block.FullName()
}

func (v *VertexVariable) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexVariable) References() []VertexReference {
	return referencesForBlock(v.block)
}

func (v *VertexVariable) Visit(mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	val, err := v.evaluator.evaluateVariable(v.block)
	if err != nil {
		return fmt.Errorf("could not evaluate variable %s: %w", v.ID(), err)
	}

	v.logger.Debugf("adding variable %s to the evaluation context", v.ID())
	key := fmt.Sprintf("var.%s", v.block.Label())
	v.evaluator.ctx.SetByDot(val, key)

	v.evaluator.AddFilteredBlocks(v.block)
	return nil
}
