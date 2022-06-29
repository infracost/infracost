package terraform

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/credentials"
)

func cloudAPI(host string, path string, token string) ([]byte, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://%s%s", host, path)
	log.Debugf("Calling Terraform Cloud API: %s", url)
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
