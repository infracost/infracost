package hcl

import (
	"fmt"
	"strings"
)

type Reference struct {
	blockType    Type
	typeLabel    string
	nameLabel    string
	remainder    []string
	key          string
	asString     string
	asJSONString string
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

	remainderIndex := 3

	if ref.blockType.removeTypeInReference && parts[0] != blockType.name {
		ref.typeLabel = parts[0]
		if len(parts) > 1 {
			ref.nameLabel = parts[1]
		}
		remainderIndex = 2
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

	if len(parts) > remainderIndex {
		ref.remainder = parts[remainderIndex:]
	}

	return &ref, nil
}

// JSONString returns the reference so that it's possible to use in the plan JSON file.
// This strips any count keys from the reference.
func (r *Reference) JSONString() string {
	if r.asJSONString != "" {
		return r.asJSONString
	}

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

	r.asJSONString = base

	return r.asJSONString
}

func (r *Reference) String() string {
	if r.asString != "" {
		return r.asString
	}

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

	r.asString = base

	return r.asString
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
