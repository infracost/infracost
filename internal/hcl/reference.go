package hcl

import (
	"fmt"
	"strings"
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
