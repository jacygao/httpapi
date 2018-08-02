package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

var (
	errMissingTimeout = errors.New("missing timeout value in the configuration")
)

// Config stores the Service configuration
type Config struct {
	// Timeout in seconds
	Timeout int
}

// DefaultConfig returns a Config struct with default values.
func DefaultConfig() Config {
	return Config{
		Timeout: 10,
	}
}

// Client wraps the http.Client worker with configuration values
type Client struct {
	Client *http.Client
}

// NewClient creates a new HTTP Requester service
func NewClient(config Config) (*Client, error) {
	client := &http.Client{}
	var err error

	// Add timeout configuration to http client
	client, err = withTimeout(config, client)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: client,
	}, nil
}

// Do wraps ctxhttp.Do using the managed http.Client
func (s *Client) Do(ctx context.Context, req *http.Request, retval interface{}, user User, authenticator Authenticator) error {
	if authenticator == nil {
		authenticator = NewDefaultAuthenticator(s.Client)
	}

	authenticatedClient, err := authenticator.AuthenticatedClient(ctx, user)
	if err != nil {
		return err
	}

	res, err := ctxhttp.Do(ctx, authenticatedClient, req)
	if err != nil {
		return err
	}

	if retval != nil {
		if err := json.NewDecoder(res.Body).Decode(retval); err != nil {
			// FIXME(jacy): Some endpoints return empty responses which causes the JSON decoder returning an EOF error
			// There may be a better way to properly handle it. For now we just skip it.
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
	}

	return nil
}

func withTimeout(config Config, client *http.Client) (*http.Client, error) {
	if config.Timeout == 0 {
		return nil, errMissingTimeout
	}
	client.Timeout = time.Second * time.Duration(config.Timeout)
	return client, nil
}
