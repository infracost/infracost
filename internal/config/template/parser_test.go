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
		name string
	}{
		{
			name: "different env dirs",
		},
		{
			name: "different env files",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDataPath := filepath.Join("./testdata", strings.ReplaceAll(tt.name, " ", "_"))
			input := filepath.Join(testDataPath, "infracost.yml.tmpl")
			golden := filepath.Join(testDataPath, "infracost.golden.yml")

			f, err := os.Open(golden)
			require.NoError(t, err)

			p := NewParser(testDataPath)

			wr := &bytes.Buffer{}
			err = p.Compile(input, wr)
			require.NoError(t, err)

			var actual interface{}
			err = yaml.NewDecoder(wr).Decode(&actual)
			require.NoError(t, err)

			var expected interface{}
			err = yaml.NewDecoder(f).Decode(&expected)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}
