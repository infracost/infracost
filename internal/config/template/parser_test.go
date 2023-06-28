package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParser_Compile(t *testing.T) {
	tests := []struct {
		name      string
		variables Variables
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDataPath := filepath.Join("./testdata", strings.ReplaceAll(tt.name, " ", "_"))
			input := filepath.Join(testDataPath, "infracost.yml.tmpl")
			golden := filepath.Join(testDataPath, "infracost.golden.yml")

			f, err := os.Open(golden)
			require.NoError(t, err)

			p := NewParser(testDataPath, tt.variables)

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
