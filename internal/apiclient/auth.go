package apiclient

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/ui"
)

// AuthClient represents a client for Infracost's authentication process.
type AuthClient struct {
	Host string
}

type callbackServerResp struct {
	err     error
	apiKey  string
	infoMsg string
}

// Login opens a browser with authentication URL and starts a HTTP server to
// wait for a callback request.
func (a AuthClient) Login(contextVals map[string]interface{}) (string, string, error) {
	state := uuid.NewString()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", "", err
	}
	port := listener.Addr().(*net.TCPAddr).Port

	q := url.Values{}
	q.Set("cli_port", fmt.Sprint(port))
	q.Set("cli_state", state)
	q.Set("cli_version", fmt.Sprintf("%v", contextVals["version"]))
	q.Set("os", fmt.Sprintf("%v", contextVals["os"]))
	q.Set("utm_source", "cli")

	startURL := fmt.Sprintf("%s/login?%s", a.Host, q.Encode())

	fmt.Println("\nIf the redirect doesn't work, use this URL:")
	fmt.Printf("\n%s\n\n", startURL)
	fmt.Printf("Waiting...\n\n")

	_ = browser.OpenURL(startURL)

	apiKey, info, err := a.startCallbackServer(listener, state)
	if err != nil {
		return "", "", err
	}

	return apiKey, info, nil
}

func (a AuthClient) startCallbackServer(listener net.Listener, generatedState string) (string, string, error) {
	shutdown := make(chan callbackServerResp, 1)

	go func() {
		defer close(shutdown)

		for {
			select {
			case <-time.After(time.Minute * 5):
				shutdown <- callbackServerResp{err: fmt.Errorf("timeout")}
				listener.Close()
				return
			}
		}
	}()

	go func() {
		_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				return
			}

			query := r.URL.Query()
			state := query.Get("cli_state")
			apiKey := query.Get("api_key")
			infoMsg := query.Get("info")
			redirectTo := query.Get("redirect_to")

			if state != generatedState {
				log.Debug("Invalid state received")
				w.WriteHeader(400)
				return
			}

			u, err := url.Parse(redirectTo)
			if err != nil {
				log.Debug("Unable to parse redirect_to URL")
				w.WriteHeader(400)
				return
			}

			origin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			if origin != a.Host {
				log.Debug("Invalid redirect_to URL")
				w.WriteHeader(400)
				return
			}

			http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
			// Flush the response, otherwise the HTTP redirect response doesn't always get sent
			// before the server shuts down.
			flusher, ok := w.(http.Flusher)
			if ok {
				flusher.Flush()
			}
			shutdown <- callbackServerResp{apiKey: apiKey, infoMsg: infoMsg}
		}))
	}()

	resp := <-shutdown

	if resp.infoMsg != "" {
		return "", resp.infoMsg, nil
	}

	if resp.apiKey == "" || resp.err != nil {
		return "", "", fmt.Errorf("Authentication failed. Please check your API token on %s", ui.LinkString("https://infracost.io"))
	}

	return resp.apiKey, "", nil
}
