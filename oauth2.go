package httpapi

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	errMissingAuthCode = errors.New("missing authencation code")
	errMissingUserID   = errors.New("missing user ID")
)

// TokenStore defines behaviours of Get and Save token.
type TokenStore interface {
	GetToken(user User) (*oauth2.Token, error)
	SaveToken(user User, token *oauth2.Token) error
}

// Oauth2Authenticator contains necessary dependencies of Oauth2Authenticator.
type Oauth2Authenticator struct {
	oauthConfig *oauth2.Config
	store       TokenStore
	httpCli     *http.Client
}

// NewOauth2Authenticator initialises a new Oauth2Authenticator instance which implements the Authenticator interface.
func NewOauth2Authenticator(config *oauth2.Config, ts TokenStore, cli *http.Client) *Oauth2Authenticator {
	return &Oauth2Authenticator{
		oauthConfig: config,
		store:       ts,
		httpCli:     cli,
	}
}

// AuthenticatedClient implements authenticator interface and provides required logic to manage tokens (access and refresh).
// If this process happens to be successful, a valid http client is returned otherwise an error is returned.
func (oa *Oauth2Authenticator) AuthenticatedClient(ctx context.Context, user User) (*http.Client, error) {
	if user == nil {
		return nil, errMissingUserID
	}

	// Add HTTP Client with transport(TLS) to context for the token-refeshing process.
	ctx = newHTTPContext(ctx, oa.httpCli)

	// GetToken should handle how tokens are retrieved.
	// This may include a caching layer for fast access backed by a persistant datastore(db).
	// Ideally this should be handled by a datastore implementation on the application level.
	authCode := user.AuthCode()
	if authCode == "" {
		return nil, errMissingAuthCode
	}
	if user.UserID() == "" {
		return nil, errMissingUserID
	}

	tok, err := oa.store.GetToken(user)
	if err != nil {
		return nil, err
	}

	// If token does not exist in TokenStore, we request a new token from auth endpoint.
	if tok == nil {
		newTok, err := oa.oauthConfig.Exchange(ctx, user.AuthCode())
		if err != nil {
			return nil, err
		}
		tok = newTok

		// Store the token in the token store
		err = oa.store.SaveToken(user, tok)
		if err != nil {
			return nil, err
		}
	}

	// Construct the http client by wrapping the appropriate innards to hook into its token-refreshing process,
	// and store refreshed access tokens into the cache/db.
	ts := oa.oauthConfig.TokenSource(ctx, tok)

	// newTokenProvider wraps TokenSource of the standard oauth2 library.
	// The wrapper allows us to add additional logic before and after the default TokenSource
	// such as communicating with a tokenStore.
	cts := newTokenProvider(user, ts, oa.store)

	return oauth2.NewClient(ctx, cts), nil
}

func newHTTPContext(ctx context.Context, c *http.Client) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, c)
}

func newTokenProvider(user User, src oauth2.TokenSource, ts TokenStore) oauth2.TokenSource {
	return &tokenProvider{
		user: user,
		pts:  src,
		ts:   ts,
	}
}

type tokenProvider struct {
	user User
	pts  oauth2.TokenSource // called when t is expired.
	ts   TokenStore
}

// Token implements the oauth2.TokenSource interface.
func (s *tokenProvider) Token() (*oauth2.Token, error) {
	t, err := s.pts.Token()
	if err != nil {
		return nil, err
	}
	if err := s.ts.SaveToken(s.user, t); err != nil {
		return nil, err
	}
	return t, nil
}
