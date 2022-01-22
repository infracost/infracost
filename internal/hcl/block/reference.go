package block

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

type Reference struct {
	blockType Type
	typeLabel string
	nameLabel string
	remainder []string
	key       string
}

func newReference(parts []string) (*Reference, error) {

	var ref Reference

	if len(parts) == 0 {
		return nil, fmt.Errorf("cannot create empty reference")
	}

	blockType, err := TypeFromRefName(parts[0])
	if err != nil {
		blockType = &TypeResource
	}

	ref.blockType = *blockType

	if ref.blockType.removeTypeInReference && parts[0] != blockType.name {
		ref.typeLabel = parts[0]
		if len(parts) > 1 {
			ref.nameLabel = parts[1]
		}
	} else {
		if len(parts) > 1 {
			ref.typeLabel = parts[1]
			if len(parts) > 2 {
				ref.nameLabel = parts[2]
			} else {
				ref.nameLabel = ref.typeLabel
				ref.typeLabel = ""
			}
		}
	}

	if strings.Contains(ref.nameLabel, "[") {
		bits := strings.Split(ref.nameLabel, "[")
		ref.nameLabel = bits[0]
		ref.key = "[" + bits[1]
	}

	if len(parts) > 3 {
		ref.remainder = parts[3:]
	}

	return &ref, nil
}

func (r *Reference) BlockType() Type {
	return r.blockType
}

func (r *Reference) TypeLabel() string {
	return r.typeLabel
}

func (r *Reference) NameLabel() string {
	return r.nameLabel
}

func (r *Reference) String() string {

	base := fmt.Sprintf("%s.%s", r.typeLabel, r.nameLabel)

	if !r.blockType.removeTypeInReference {
		base = r.blockType.Name()
		if r.typeLabel != "" {
			base += "." + r.typeLabel
		}
		if r.nameLabel != "" {
			base += "." + r.nameLabel
		}
	}

	if r.key != "" {
		base += r.key
	}

	for _, rem := range r.remainder {
		base += "." + rem
	}

	return base
}

func (r *Reference) RefersTo(b Block) bool {
	if r.BlockType() != b.Reference().BlockType() {
		return false
	}
	if r.TypeLabel() != b.Reference().TypeLabel() {
		return false
	}
	if r.NameLabel() != b.Reference().NameLabel() {
		return false
	}
	if r.Key() != "" && r.Key() != b.Reference().Key() {
		return false
	}
	return true
}

func (r *Reference) SetKey(key cty.Value) {
	switch key.Type() {
	case cty.Number:
		f := key.AsBigFloat()
		f64, _ := f.Float64()
		r.key = fmt.Sprintf("[%d]", int(f64))
	case cty.String:
		r.key = fmt.Sprintf("[%q]", key.AsString())
	}
}

func (r *Reference) Key() string {
	return r.key
}
