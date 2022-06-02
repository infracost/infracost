package apiclient

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
)

// AuthClient represents a client for Infracost's authentication process.
type AuthClient struct {
	Host string
}

// Login opens a browser with authentication URL and starts a HTTP server to
// wait for a callback request.
func (a AuthClient) Login() (string, error) {
	state := uuid.NewString()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	port := listener.Addr().(*net.TCPAddr).Port

	q := url.Values{}
	q.Set("cli_port", fmt.Sprint(port))
	q.Set("cli_state", state)
	q.Set("utm_source", "cli")

	startURL := fmt.Sprintf("%s/login?%s", a.Host, q.Encode())

	fmt.Println("\nIf the redirect doesn't work, use this URL:")
	fmt.Printf("\n%s\n\n", startURL)
	fmt.Printf("Waiting...\n\n")

	_ = browser.OpenURL(startURL)

	apiKey, err := a.startCallbackServer(listener, state)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (a AuthClient) startCallbackServer(listener net.Listener, generatedState string) (string, error) {
	apiKey := ""

	err := http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		defer listener.Close()

		query := r.URL.Query()
		state := query.Get("cli_state")
		apiKey = query.Get("api_key")

		if apiKey == "" || state != generatedState {
			w.WriteHeader(400)
			return
		}
	}))

	if !errors.Is(err, net.ErrClosed) {
		return "", err
	}

	if apiKey == "" {
		return "", errors.New("Unable to receive API key")
	}

	return apiKey, nil
}
