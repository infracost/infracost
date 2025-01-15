package modules

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpPublicModuleChecker struct {
	client *http.Client
}

func NewHttpPublicModuleChecker() *HttpPublicModuleChecker {
	return &HttpPublicModuleChecker{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// IsPublic checks if a module is public by making a HEAD request to the module address
// and checking if the response status code is 200.
func (h *HttpPublicModuleChecker) IsPublicModule(moduleAddr string) (bool, error) {
	if strings.HasPrefix(moduleAddr, "git@") {
		// We don't support git@ urls
		return false, nil
	}

	u := strings.TrimPrefix(moduleAddr, "git::")

	parsedUrl, err := url.Parse(u)
	if err != nil {
		return false, err
	}

	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = "https"
	}

	if parsedUrl.User != nil {
		return false, nil
	}

	req, err := http.NewRequest("HEAD", parsedUrl.String(), nil)
	if err != nil {
		return false, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
