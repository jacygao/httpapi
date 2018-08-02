package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func TestClientWithDefaultAuthenticator(t *testing.T) {
	// Creating a testing http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer ts.Close()

	config := Config{
		Timeout: 10,
	}
	requester, err := NewClient(config)
	if err != nil {
		t.Fatalf(err.Error())
	}

	provider := NewDefaultAuthenticator(requester.Client)
	req, err := http.NewRequest("POST", ts.URL, nil)
	if err != nil {
		t.Fatalf("Problem creating request. Error: %s", err.Error())
	}

	var res interface{}
	if err := requester.Do(context.Background(), req, &res, nil, provider); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestClientWithOauth2(t *testing.T) {
	// Creating a testing http server to handle token validation
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "mock2", "refresh_token": "refreshmock", "scope": "user", "token_type": "bearer", "expires_in": 86400}`))
	}))
	defer ts.Close()

	config := Config{
		Timeout: 10,
	}
	requester, err := NewClient(config)
	if err != nil {
		t.Fatalf(err.Error())
	}

	oauthConfig := &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			AuthURL:  ts.URL,
			TokenURL: ts.URL,
		},
	}

	memstore := NewMemStore()
	provider := NewOauth2Authenticator(oauthConfig, memstore, requester.Client)

	// According to the Oauth2 3-legged authentication workflow, a valid serverCode will be passed to the server by the client.
	mockSrvCode := "1234"
	mockUserID := "mockID"
	user := NewOauth2User(mockUserID, mockSrvCode)

	mockToken := &oauth2.Token{
		AccessToken:  "mock",
		TokenType:    "Bearer",
		RefreshToken: "refreshmock",
		// Setting expiry to current time means it will expire by the time we make a request.
		Expiry: time.Now().Local(),
	}

	if err := memstore.SaveToken(user, mockToken); err != nil {
		t.Fatalf(err.Error())
	}

	req, err := http.NewRequest("POST", ts.URL, nil)
	if err != nil {
		t.Fatalf("Problem creating request. Error: %s", err.Error())
	}

	var res *oauth2.Token
	if err := requester.Do(context.Background(), req, &res, user, provider); err != nil {
		t.Fatalf(err.Error())
	}

	// httpapi should automatically update any expired token with the refresh token.
	// So here we are checking if the access token has been updated.
	newtok, err := memstore.GetToken(user)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if newtok.AccessToken != "mock2" {
		t.Fatalf("Expected updated access token %s but received %s \n", "mock2", newtok.AccessToken)
	}

	if newtok.RefreshToken != mockToken.RefreshToken {
		t.Fatalf("Expected static refresh token %s but received %s \n", mockToken.RefreshToken, newtok.RefreshToken)
	}

	if time.Since(newtok.Expiry) > 0 {
		t.Fatalf("Expected extended expiry but received %s \n", newtok.Expiry)
	}
}

// Oauth2User implement the User interface to provide required data for authentication
type Oauth2User struct {
	id       string
	authCode string
}

// NewOauth2User creates a new Oauth2User struct
func NewOauth2User(id string, authcode string) *Oauth2User {
	return &Oauth2User{
		id:       id,
		authCode: authcode,
	}
}

// UserID returns an Oauth2User's ID
func (u *Oauth2User) UserID() string {
	return u.id
}

// AuthCode returns an Oauth2User's authentciation code
func (u *Oauth2User) AuthCode() string {
	return u.authCode
}
