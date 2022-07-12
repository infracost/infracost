package modules

import (
	"io"
	"testing"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/sirupsen/logrus"
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
		"module-d": {
			Key:    "module-c",
			Source: "app.terraform.io/infracost/ec2-instance/aws",
			Dir:    ".infracost/module-c",
		},
		"submodule-a": {
			Key:     "submodule-a",
			Source:  "registry.terraform.io/namespace/module-a/aws//submodule/path",
			Version: "1.0.0",
			Dir:     ".infracost/module-a/submodule/path",
		},
		"submodule-b": {
			Key:    "submodule-a",
			Source: "git::https://github.com/namespace/module-b.git//submodule/path?v=0.5.0",
			Dir:    ".infracost/module-b/submodule/path",
		},
	}

	l := logrus.New()
	l.SetOutput(io.Discard)
	logger := logrus.NewEntry(l)

	cache := &Cache{
		keyMap: keyMap,
		disco:  NewDisco(nil, logrus.NewEntry(logrus.New())),
		logger: logger,
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
		{"module-d", &tfconfig.ModuleCall{Source: "app.terraform.io/infracost/ec2-instance/aws"}, keyMap["module-d"], ""},
		{"submodule-a", &tfconfig.ModuleCall{Source: "registry.terraform.io/namespace/module-a/aws//submodule/path", Version: ">=1.0"}, keyMap["submodule-a"], ""},
		{"submodule-a", &tfconfig.ModuleCall{Source: "namespace/module-a/aws//submodule/path", Version: ">=1.0"}, keyMap["submodule-a"], ""},
		{"submodule-a", &tfconfig.ModuleCall{Source: "registry.terraform.io/namespace/module-a/aws//submodule/path", Version: ">=2.0"}, nil, "version constraint doesn't match"},
		{"submodule-a", &tfconfig.ModuleCall{Source: "registry.terraform.io/different-namespace/module-a-/aws//submodule/path", Version: "1.0.0"}, nil, "source has changed"},
		{"submodule-b", &tfconfig.ModuleCall{Source: "git::https://github.com/namespace/module-b.git//submodule/path?v=0.5.0"}, keyMap["submodule-b"], ""},
		{"submodule-b", &tfconfig.ModuleCall{Source: "git::https://github.com/namespace/module-b.git//submodule/path?v=0.6.0"}, nil, "source has changed"},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			actual, err := cache.lookupModule(test.key, test.moduleCall)

			actualErr := ""
			if err != nil {
				actualErr = err.Error()
			}
			assert.Equal(t, test.expectedError, actualErr)
			assert.Equal(t, test.expected, actual)
		})
	}
}
