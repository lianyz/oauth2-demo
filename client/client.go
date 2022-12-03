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
	"flag"
	"fmt"
	"github.com/lianyz/oauth2-demo/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	authentication "k8s.io/api/authentication/v1beta1"
)

type userInfo struct {
	clientId  string `json:"client_id"`
	expiresIn int    `json:"expires_in"`
	userId    string `json:"user_id"`
}

const (
	state         = "xyz"
	challengeCode = "s256example"
	port          = "9094"
)

var (
	authServerURL string
	clientId      string
	clientSecret  string
	clientAddr    string
	config        *oauth2.Config
	globalToken   *oauth2.Token // Non-concurrent security
)

func init() {
	flag.StringVar(&clientId, "id", "123456", "client id")
	flag.StringVar(&clientSecret, "secret", "111111", "client secret")
	flag.StringVar(&clientAddr, "addr", "http://localhost:9094", "client addr")
	flag.StringVar(&authServerURL, "server", "http://localhost:9096", "auth server addr")
}

func main() {

	initConfig()

	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/oauth2", tokenHandler)
	http.HandleFunc("/refresh", refreshHandler)
	http.HandleFunc("/try", tryHandler)
	http.HandleFunc("/pwd", pwdHandler)
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	http.HandleFunc("/k8s-authn-webhook", webhookHandler)

	log.Printf("Client is running at %s. Please open %s", clientAddr, clientAddr)

	items := strings.Split(clientAddr, ":")
	log.Fatal(http.ListenAndServe(":"+items[2], nil))
}

func initConfig() {
	flag.Parse()
	config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"all"},
		RedirectURL:  clientAddr + "/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:  authServerURL + "/oauth/authorize",
			TokenURL: authServerURL + "/oauth/token",
		},
	}

	fmt.Printf("ClientId: %s ClientSecret: %s ClientAddr: %s\n", clientId, clientSecret, clientAddr)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest("/", r)

	url := config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256(challengeCode)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))

	// 向Authorization Server发送请求
	// /oauth/authorize?client_id=222222
	// &code_challenge=Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=
	// &code_challenge_method=S256
	// &redirect_uri=http://localhost:9094/oauth2
	// &response_type=code
	// &scope=all
	// &state=xyz
	redirect(w, r, url)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest("token", r)
	r.ParseForm()
	state := r.Form.Get("state")
	if state != state {
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
	log.Printf("token: %v", token)
	encodeToken(w, token)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest("refresh", r)
	if globalToken == nil {
		redirect(w, r, "/")
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
	utils.LogRequest("try", r)
	if globalToken == nil {
		redirect(w, r, "/")
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

// webhookHandler serve as k8s authentication webhook
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest("webhook", r)

	decoder := json.NewDecoder(r.Body)
	var tokenReview authentication.TokenReview
	err := decoder.Decode(&tokenReview)
	if err != nil {
		log.Println("[Error] Decode", err.Error())
		writeTokenReviewStatusFailed(w, http.StatusBadRequest)

		return
	}

	res, err := http.Get(fmt.Sprintf("%s/test?access_token=%s",
		authServerURL, tokenReview.Spec.Token))
	if err != nil {
		log.Println("[Error] authorize", err.Error())
		writeTokenReviewStatusFailed(w, http.StatusUnauthorized)

		return
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var user userInfo
	err = decoder.Decode(user)
	if err != nil {
		log.Println("[Error] Parse User Info ", err.Error())
		writeTokenReviewStatusFailed(w, http.StatusUnauthorized)

		return
	}

	log.Printf("[Success] login as %v", user)
	writeTokenReviewStatusOK(w, user)
}

func pwdHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest("pwd", r)
	token, err := config.PasswordCredentialsToken(context.Background(), "test", "test")
	if err != nil {
		internalServerError(w, err.Error())
		return
	}

	globalToken = token
	encodeToken(w, token)
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest("client", r)
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

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	filePath, err := utils.GetRunPath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fullFileName := filepath.Join(filePath, "favicon.ico")
	http.ServeFile(w, r, fullFileName)
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	utils.LogRedirect(r.URL.String(), url)
	http.Redirect(w, r, url, http.StatusFound)
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

func writeTokenReviewStatusFailed(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "authentication.k8s.io/v1beta1",
		"kind":       "TokenReview",
		"status": authentication.TokenReviewStatus{
			Authenticated: false,
		},
	})
}

func writeTokenReviewStatusOK(w http.ResponseWriter, user userInfo) {
	w.WriteHeader(http.StatusOK)
	tokenReviewStatus := authentication.TokenReviewStatus{
		Authenticated: true,
		User: authentication.UserInfo{
			Username: user.userId,
			UID:      user.userId,
		},
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "authentication.k8s.io/v1beta1",
		"kind":       "TokenReview",
		"status":     tokenReviewStatus,
	})
}

func encodeToken(w http.ResponseWriter, token *oauth2.Token) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(token)
}
