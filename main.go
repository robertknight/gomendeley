package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"crypto/rand"
	"code.google.com/p/goauth2/oauth"
)

type ClientId struct {
	ClientId     string
	ClientSecret string
}

type MendeleyProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`

	Photo struct {
		Original string `json:"original"`
	} `json:"photo"`
}

const mendeleyEndpoint = "https://mix.mendeley.com"

var oauthConfig = &oauth.Config{
	// this specifies the data we want access to,
	// at the time of writing the only supported value
	// is 'all'
	Scope: "all",

	// this is the URL we redirect to in order to authenticate the user
	AuthURL:  "https://mix.mendeley.com/oauth/authorize",
	TokenURL: "https://mix.mendeley.com/oauth/token",

	// the 'Redirect URI' listed for the application at
	// http://dev.mendeley.com/html/yourapps.html
	RedirectURL: "http://localhost:8080",
}

// a cache which stores access tokens for authenticated users
var tokens = map[string]oauth.Token{}

// generate a new key to identify the user to
// this web server
func generateKey() string {
	buffer := make([]byte, 16)
	rand.Read(buffer)
	return base64.StdEncoding.EncodeToString(buffer)
}

var ErrWillAuthenticate = errors.New("authentication required")

// Set a cookie to identify the user with this web server.
// If the user has already been allocated an ID this will return
// the existing ID
func setUserCookie(w http.ResponseWriter, r *http.Request) string {
	idCookie, err := r.Cookie("Key")
	var id string
	if err == http.ErrNoCookie {
		id = generateKey()
		w.Header().Add("Set-Cookie", "Key="+id)
	} else {
		id = idCookie.Value
	}
	return id
}

// Retrieve the OAuth access credentials for the current user.
// If the user has not yet authenticated with Mendeley, this returns an error
// and redirects the user to the Mendeley login page.
//
// If the user is successfully authenticated, this returns an http.Client that
// can be used to make requests against the Mendeley API
func authenticateUser(w http.ResponseWriter, r *http.Request) (*http.Client, error) {
	session := &oauth.Transport{Config: oauthConfig}
	id := setUserCookie(w, r)
	token, ok := tokens[id]
	if !ok {
		if len(r.FormValue("code")) > 0 {
			// the user successfully authenticated, get an access token
			// for future requests
			_, err := session.Exchange(r.FormValue("code"))
			if err != nil {
				return nil, err
			}
			tokens[id] = *session.Token
		} else {
			// redirect to Mendeley API auth page
			http.Redirect(w, r, session.Config.AuthCodeURL(""), http.StatusFound)
			return nil, ErrWillAuthenticate
		}
	} else {
		session.Token = &token
	}

	return session.Client(), nil
}

func mendeleyApiRequest(client *http.Client, resource string, result interface{}) error {
	resp, err := client.Get(mendeleyEndpoint + resource)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, result)
	return err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	client, err := authenticateUser(w, r)
	if err == ErrWillAuthenticate {
		return
	} else if err != nil {
		fmt.Fprintf(w, "Authentication failed: %v\n", err)
		return
	}

	var profile MendeleyProfile
	err = mendeleyApiRequest(client, "/profiles/me", &profile)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "Hello %s %s - Welcome to Mendeley :)", profile.FirstName, profile.LastName)
}

func main() {
	configFile, err := os.Open("client_config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "client_config.json file not found. See the README\n")
		os.Exit(1)
	}
	content, _ := ioutil.ReadAll(configFile)
	var clientId ClientId
	err = json.Unmarshal(content, &clientId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse client config file: %v\n", err)
		os.Exit(1)
	}

	oauthConfig.ClientId = clientId.ClientId
	oauthConfig.ClientSecret = clientId.ClientSecret

	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}
