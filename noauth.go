package httpapi

import (
	"net/http"

	"golang.org/x/net/context"
)

// DefaultAuthenticator is a noauth implementation of the Authenticator interface
type DefaultAuthenticator struct {
	httpCli *http.Client
}

// NewDefaultAuthenticator initialises a new DefaultAuthenticator instance which implements the Authenticator interface.
func NewDefaultAuthenticator(cli *http.Client) *DefaultAuthenticator {
	return &DefaultAuthenticator{
		httpCli: cli,
	}
}

// AuthenticatedClient implements ClientProvider interface and provides required logic to manage authentication.
// If this process happens to be successful, a valid http client is returned otherwise an error is returned.
func (da *DefaultAuthenticator) AuthenticatedClient(ctx context.Context, user User) (*http.Client, error) {
	return da.httpCli, nil
}
