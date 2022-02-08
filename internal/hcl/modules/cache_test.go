package modules

import (
	"testing"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/stretchr/testify/assert"
)

func TestLookupModule(t *testing.T) {
	keyMap := map[string]*ManifestModule{
		"module-a": {
			Key:     "module-a",
			Source:  "registry.terraform.io/namespace/module-a/aws",
			Version: "1.0.0",
			Dir:     ".infracost/module-a",
		},
		"module-b": {
			Key:    "module-b",
			Source: "git::https://github.com/namespace/module-b.git?v=0.5.0",
			Dir:    ".infracost/module-b",
		},
	}
	cache := &Cache{
		keyMap: keyMap,
	}

	tests := []struct {
		key           string
		moduleCall    *tfconfig.ModuleCall
		expected      *ManifestModule
		expectedError string
	}{
		{"module-a", &tfconfig.ModuleCall{Source: "registry.terraform.io/namespace/module-a/aws", Version: ">=1.0"}, keyMap["module-a"], ""},
		{"module-a", &tfconfig.ModuleCall{Source: "namespace/module-a/aws", Version: ">=1.0"}, keyMap["module-a"], ""},
		{"module-a", &tfconfig.ModuleCall{Source: "registry.terraform.io/namespace/module-a/aws", Version: ">=2.0"}, nil, "version constraint doesn't match"},
		{"module-a", &tfconfig.ModuleCall{Source: "registry.terraform.io/different-namespace/module-a-/aws", Version: "1.0.0"}, nil, "source has changed"},
		{"module-b", &tfconfig.ModuleCall{Source: "git::https://github.com/namespace/module-b.git?v=0.5.0"}, keyMap["module-b"], ""},
		{"module-b", &tfconfig.ModuleCall{Source: "git::https://github.com/namespace/module-b.git?v=0.6.0"}, nil, "source has changed"},
		{"module-c", &tfconfig.ModuleCall{Source: "git::https://github.com/namespace/module-c.git?v=0.6.0"}, nil, "not in cache"},
	}

	for _, test := range tests {
		actual, err := cache.lookupModule(test.key, test.moduleCall)

		actualErr := ""
		if err != nil {
			actualErr = err.Error()
		}
		assert.Equal(t, test.expectedError, actualErr)
		assert.Equal(t, test.expected, actual)
	}
}
