package httpapi

import (
	"net/http"

	"golang.org/x/net/context"
)

// Requester defines methods that perform HTTP requests.
type Requester interface {
	// Do performs a single http request.
	// user is optional. Passing nil if the authenticator does not require user info.
	// authenticator is optional. If nil is passed the function will make request with the default http client.
	Do(ctx context.Context, req *http.Request, retval interface{}, user User, authenticator Authenticator) error
}
