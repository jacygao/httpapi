package httpapi

import (
	"net/http"

	"golang.org/x/net/context"
)

// Authenticator defines required behaviour(s) of creating authenticated http.Client instances.
type Authenticator interface {
	AuthenticatedClient(ctx context.Context, user User) (*http.Client, error)
}

// User defines required user data for authentication.
// This interface is open for extension to work with new authenticators
type User interface {
	// UserID returns a user's unique ID
	UserID() string
	// AuthCode returns a user's authentication code.
	// This code is usually used for authentication through external sources such as Oauth.
	AuthCode() string
}
