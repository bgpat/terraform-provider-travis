package travis

import (
	"errors"
	"net/http"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shuheiktgw/go-travis"
)

// Client is an API client for Travis CI.
type Client struct {
	*travis.Client
}

// NewClient returns an API client object.
func NewClient(url, token string) *Client {
	client := &Client{
		Client: travis.NewClient(url, token),
	}
	client.HTTPClient = &http.Client{Transport: &roundTripper{base: http.DefaultTransport}}
	return client
}

type roundTripper struct {
	base http.RoundTripper
	mu   sync.Mutex
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		r.mu.Lock()
		defer r.mu.Unlock()
	}
	tflog.Debug(req.Context(), "request travis API", map[string]interface{}{
		"method": req.Method,
		"url":    req.URL.String(),
	})
	return r.base.RoundTrip(req)
}

func isNotFound(err error) bool {
	var errResp *travis.ErrorResponse
	return errors.As(err, &errResp) && errResp.ErrorType == "not_found"
}
