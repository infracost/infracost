package hcl

import (
	"fmt"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/schema"
)

type ResourceData struct {
	Block *Block
	Data  *schema.ResourceData
}

type VertexResource struct {
	moduleConfigs *ModuleConfigs
	logger        zerolog.Logger
	block         *Block
	expanded      Blocks
}

func (v *VertexResource) ID() string {
	return v.block.FullName()
}

func (v *VertexResource) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexResource) References() []VertexReference {
	return v.block.VerticesReferenced()
}

func (v *VertexResource) Visit(mutex *sync.Mutex) (interface{}, error) {
	mutex.Lock()
	defer mutex.Unlock()

	moduleInstances := v.moduleConfigs.Get(v.block.ModuleAddress())
	if len(moduleInstances) == 0 {
		return nil, fmt.Errorf("no module instances found for module address %q", v.block.ModuleAddress())
	}

	for _, moduleInstance := range moduleInstances {
		e := moduleInstance.evaluator
		blockInstance := e.module.Blocks.FindLocalName(v.block.LocalName())

		if blockInstance == nil {
			return nil, fmt.Errorf("could not find block %q in module %q", v.ID(), moduleInstance.name)
		}

		err := v.evaluate(e, blockInstance)
		if err != nil {
			return nil, fmt.Errorf("could not evaluate resource block %q", v.ID())
		}

		expanded, err := v.expand(e, blockInstance)
		if err != nil {
			return nil, fmt.Errorf("could not expand resource block %q", v.ID())
		}

		e.AddFilteredBlocks(expanded...)
		v.expanded = append(v.expanded, expanded...)
	}

	return nil, nil
}

func (v *VertexResource) TransformToSchemaResources() []ResourceData {
	var resources []ResourceData

	for _, block := range v.expanded {
		data := v.transformToSchemaResource(block)
		if data == nil {
			continue
		}

		resources = append(resources, ResourceData{
			Block: block,
			Data:  data,
		})

	}

	return resources
}

func (v *VertexResource) transformToSchemaResource(b *Block) *schema.ResourceData {
	out, err := b.MarshalValuesJSON()
	if err != nil {
		v.logger.Debug().Err(err).Msgf("could not marshal block values for resource %q, removing from output", b.FullName())
		return nil
	}

	label := b.TypeLabel()
	providerName := strings.Split(label, "_")[0]

	details := b.CallDetails()
	valDetails, _ := jsoniter.Marshal(details)
	return &schema.ResourceData{
		Type:         label,
		ProviderName: providerName,
		Address:      b.FullName(),
		// Tags are built further up the program execution at the hcl_provider level.
		Tags:          nil,
		RawValues:     gjson.ParseBytes(out.Data),
		ReferencesMap: map[string][]*schema.ResourceData{},
		Metadata: map[string]gjson.Result{
			"filename": {
				Type: gjson.String,
				Raw:  b.Filename,
				Str:  b.Filename,
			},
			"startLine": {
				Type: gjson.Number,
				Raw:  fmt.Sprintf("%d", b.StartLine),
				Num:  float64(b.StartLine),
			},
			"endLine": {
				Type: gjson.Number,
				Raw:  fmt.Sprintf("%d", b.EndLine),
				Num:  float64(b.EndLine),
			},
			"calls": gjson.ParseBytes(valDetails),
			"checksum": {
				Type: gjson.String,
				Raw:  out.CheckSum,
				Str:  out.CheckSum,
			},
		},
	}

}

func (v *VertexResource) evaluate(e *Evaluator, b *Block) error {
	if len(b.Labels()) < 2 {
		return fmt.Errorf("resource block %s has no label", v.ID())
	}

	var existingVals map[string]cty.Value
	existingCtx := e.ctx.Get(b.TypeLabel())
	if !existingCtx.IsNull() {
		existingVals = existingCtx.AsValueMap()
	} else {
		existingVals = make(map[string]cty.Value)
	}

	val := e.evaluateResource(b, existingVals)

	v.logger.Debug().Msgf("adding resource %s to the evaluation context", v.ID())
	e.ctx.SetByDot(val, b.TypeLabel())

	return nil
}

func (v *VertexResource) expand(e *Evaluator, b *Block) ([]*Block, error) {
	expanded := []*Block{b}
	expanded = e.expandBlockCounts(expanded)
	expanded = e.expandBlockForEaches(expanded)
	expanded = e.expandDynamicBlocks(expanded...)

	b.expanded = true

	return expanded, nil
}
