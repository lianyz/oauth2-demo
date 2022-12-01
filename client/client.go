/*
@Time : 2022/12/1 11:52
@Author : lianyz
@Description :
*/

package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	authServerURL = "http://localhost:9096"
	stateCode     = "xyz"
	challengeCode = "s256example"
	port          = "9094"
)

var (
	config = oauth2.Config{
		ClientID:     "222222",
		ClientSecret: "22222222",
		Scopes:       []string{"all"},
		RedirectURL:  "http://localhost:9094/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:  authServerURL + "/oauth/authorize",
			TokenURL: authServerURL + "/oauth/token",
		},
	}
	globalToken *oauth2.Token // Non-concurrent security
)

func main() {
	http.HandleFunc("/", challengeHandler)
	http.HandleFunc("/oauth2", tokenHandler)
	http.HandleFunc("/refresh", refreshHandler)
	http.HandleFunc("/try", tryHandler)
	http.HandleFunc("/pwd", pwdHandler)
	http.HandleFunc("/client", clientHandler)

	log.Printf("Client is running at :%s port. Please open http://localhost:%s", port, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func challengeHandler(w http.ResponseWriter, r *http.Request) {
	url := config.AuthCodeURL(stateCode,
		oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256(challengeCode)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	http.Redirect(w, r, url, http.StatusFound)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	state := r.Form.Get("state")
	if state != stateCode {
		badRequestError(w, "State invalid")
		return
	}
	code := r.Form.Get("code")
	if code == "" {
		badRequestError(w, "Code not found")
		return
	}
	token, err := config.Exchange(context.Background(), code, oauth2.SetAuthURLParam("code_verifier", challengeCode))
	if err != nil {
		internalServerError(w, err.Error())
		return
	}
	globalToken = token

	encodeToken(w, token)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if globalToken == nil {
		redirectToChallenge(w, r)
		return
	}

	globalToken.Expiry = time.Now()
	token, err := config.TokenSource(context.Background(), globalToken).Token()
	if err != nil {
		internalServerError(w, err.Error())
		return
	}

	globalToken = token
	encodeToken(w, token)
}

func tryHandler(w http.ResponseWriter, r *http.Request) {
	if globalToken == nil {
		redirectToChallenge(w, r)
		return
	}

	res, err := http.Get(fmt.Sprintf("%s/test?access_token=%s",
		authServerURL, globalToken.AccessToken))
	if err != nil {
		badRequestError(w, err.Error())
		return
	}
	defer res.Body.Close()
	io.Copy(w, res.Body)
}

func pwdHandler(w http.ResponseWriter, r *http.Request) {
	token, err := config.PasswordCredentialsToken(context.Background(), "test", "test")
	if err != nil {
		internalServerError(w, err.Error())
		return
	}

	globalToken = token
	encodeToken(w, token)
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	cfg := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.Endpoint.TokenURL,
	}

	token, err := cfg.Token(context.Background())
	if err != nil {
		internalServerError(w, err.Error())
		return
	}

	encodeToken(w, token)
}

func redirectToChallenge(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(s256[:])
}

func badRequestError(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusBadRequest)
}

func internalServerError(w http.ResponseWriter, err string) {
	http.Error(w, err, http.StatusInternalServerError)
}

func encodeToken(w http.ResponseWriter, token *oauth2.Token) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(token)
}
