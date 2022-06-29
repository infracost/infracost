package hcl

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
	} else if len(parts) > 1 {
		ref.typeLabel = parts[1]
		if len(parts) > 2 {
			ref.nameLabel = parts[2]
		} else {
			ref.nameLabel = ref.typeLabel
			ref.typeLabel = ""
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

func (r *Reference) SetKey(key cty.Value) {
	if !key.IsKnown() {
		return
	}

	if key.IsNull() {
		return
	}

	switch key.Type() {
	case cty.Number:
		f := key.AsBigFloat()
		f64, _ := f.Float64()
		r.key = fmt.Sprintf("[%d]", int(f64))
	case cty.String:
		r.key = fmt.Sprintf("[%q]", key.AsString())
	}
}

// JSONString returns the reference so that it's possible to use in the plan JSON file.
// This strips any count keys from the reference.
func (r *Reference) JSONString() string {
	base := fmt.Sprintf("%s.%s", r.typeLabel, r.nameLabel)

	if !r.blockType.removeTypeInReference {
		base = r.blockType.Name()
		if r.blockType.Name() == "variable" {
			base = "var"
		}

		if r.typeLabel != "" {
			base += "." + r.typeLabel
		}
		if r.nameLabel != "" {
			base += "." + r.nameLabel
		}
	}

	return base
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

type Type struct {
	name                  string
	refName               string
	removeTypeInReference bool
}

func (t Type) Name() string {
	return t.name
}

func (t Type) ShortName() string {
	if t.refName != "" {
		return t.refName
	}
	return t.name
}

var TypeData = Type{
	name: "data",
}

var TypeResource = Type{
	name:                  "resource",
	removeTypeInReference: true,
}

var TypeVariable = Type{
	name:    "variable",
	refName: "var",
}

var TypeLocal = Type{
	name:    "locals",
	refName: "local",
}

var TypeProvider = Type{
	name: "provider",
}

var TypeOutput = Type{
	name: "output",
}

var TypeModule = Type{
	name: "module",
}

var TypeTerraform = Type{
	name: "terraform",
}

var ValidTypes = []Type{
	TypeData,
	TypeLocal,
	TypeModule,
	TypeOutput,
	TypeProvider,
	TypeResource,
	TypeTerraform,
	TypeVariable,
}

func TypeFromRefName(name string) (*Type, error) {
	for _, valid := range ValidTypes {
		if valid.refName == name || (valid.refName == "" && valid.name == name) {
			return &valid, nil
		}
	}
	return nil, fmt.Errorf("block type not found")
}
