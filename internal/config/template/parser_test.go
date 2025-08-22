package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParser_Compile(t *testing.T) {
	tests := []struct {
		name      string
		variables Variables
		config    *config.Config
	}{
		{
			name: "different env dirs",
		},
		{
			name: "different env files",
		},
		{
			name: "external dirs",
		},
		{
			name: "include directory based on file",
		},
		{
			name: "with string matching functions",
		},
		{
			name: "with string manipulation functions",
		},
		{
			name: "with top level template data",
			variables: Variables{
				Branch:     "test",
				BaseBranch: "master",
			},
		},
		{
			name: "with rel paths",
		},
		{
			name: "with list",
		},
		{
			name: "with is dir",
		},
		{
			name: "with parse functions",
		},
		{
			name: "with is production",
			config: &config.Config{
				Configuration: config.Configuration{
					ProductionFilters: []config.ProductionFilter{
						{
							Type:    "PROJECT",
							Value:   "test",
							Include: true,
						},
						{
							Type:    "PROJECT",
							Value:   "test2",
							Include: false,
						},
						{
							Type:    "PROJECT",
							Value:   "foo*",
							Include: true,
						},
						{
							Type:    "PROJECT",
							Value:   "foo*bar",
							Include: true,
						},
						{
							Type:    "PROJECT",
							Value:   "foo1",
							Include: false,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDataPath := filepath.Join("./testdata", strings.ReplaceAll(tt.name, " ", "_"))
			input := filepath.Join(testDataPath, "infracost.yml.tmpl")
			golden := filepath.Join(testDataPath, "infracost.golden.yml")

			f, err := os.Open(golden)
			require.NoError(t, err)

			p := NewParser(testDataPath, tt.variables, tt.config)

			wr := &bytes.Buffer{}
			err = p.CompileFromFile(input, wr)
			require.NoError(t, err)

			contents, err := os.ReadFile(input)
			require.NoError(t, err)

			wrr := &bytes.Buffer{}
			err = p.Compile(string(contents), wrr)
			require.NoError(t, err)

			var actualFileOutput interface{}
			err = yaml.NewDecoder(wr).Decode(&actualFileOutput)
			require.NoError(t, err)

			var actualStringOutput interface{}
			err = yaml.NewDecoder(wrr).Decode(&actualStringOutput)
			require.NoError(t, err)

			var expected interface{}
			err = yaml.NewDecoder(f).Decode(&expected)
			require.NoError(t, err)
			assert.Equal(t, expected, actualFileOutput)
			assert.Equal(t, expected, actualStringOutput)
		})
	}
}
