package terraform

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/credentials"
	"github.com/infracost/infracost/internal/logging"
)

func cloudAPI(host string, path string, token string) ([]byte, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://%s%s", host, path)
	logging.Logger.Debug().Msgf("Calling Terraform Cloud API: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return []byte{}, credentials.ErrInvalidCloudToken
	} else if resp.StatusCode != 200 {
		return []byte{}, errors.Errorf("invalid response from Terraform remote: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
