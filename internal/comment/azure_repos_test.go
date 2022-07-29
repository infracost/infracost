package comment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildAzureAPIURL(t *testing.T) {
	tests := []struct {
		repoURL string
		want    string
	}{
		{
			repoURL: "https://SG-GDI-CTO-PublicCloud@dev.azure.com/SG-GDI-CTO-PublicCloud/CloudSolutions%20Playground/_git/CloudSolutions%20Playground",
			want:    "https://dev.azure.com/SG-GDI-CTO-PublicCloud/CloudSolutions%20Playground/_apis/git/repositories/CloudSolutions%20Playground/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.repoURL, func(t *testing.T) {
			got, err := buildAzureAPIURL(tt.repoURL)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
