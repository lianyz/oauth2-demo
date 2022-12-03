/*
@Time : 2022/12/1 08:37
@Author : lianyz
@Description :
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-session/session"
	"github.com/lianyz/oauth2-demo/utils"
	"path/filepath"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"

	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

var (
	dump   bool
	id     string
	secret string
	domain string
	port   int
)

func init() {
	flag.BoolVar(&dump, "d", true, "Dump requests and responses")
	flag.StringVar(&id, "i", "222222", "The client id being passed in")
	flag.StringVar(&secret, "s", "22222222", "The client secret being passed in")
	flag.StringVar(&domain, "r", "http://localhost:9094", "The domain of the redirect url")
	flag.IntVar(&port, "p", 9096, "the base port for the server")
}

func main() {
	flag.Parse()
	if dump {
		log.Println("Dumping requests")
	}
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	manager.MapAccessGenerate(generates.NewAccessGenerate())

	clientStore := store.NewClientStore()
	manager.MapClientStorage(clientStore)

	clientStore.Set(id, &models.Client{
		ID:     id,
		Secret: secret,
		Domain: domain,
	})

	srv := server.NewServer(server.NewConfig(), manager)

	srv.SetPasswordAuthorizationHandler(passwordAuthorizeHandler)
	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (res *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(res *errors.Response) {
		log.Println("Response Error:", res.Error.Error())
	})

	addHandler("/login", login, nil)
	addHandler("/auth", auth, nil)
	addHandler("/oauth/authorize", authorize, srv)
	addHandler("/oauth/token", token, srv)
	addHandler("/test", test, srv)

	endpoint := fmt.Sprintf("%s:%d", "http://localhost", port)
	log.Printf("Server is running at %d port.\n", port)
	log.Printf("Point your OAuth client Auth endpoint to %s%s", endpoint, "/oauth/authorize")
	log.Printf("Point your OAuth client Token endpoint to %s%s", endpoint, "/oauth/token")

	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func passwordAuthorizeHandler(ctx context.Context, clientID, username, password string) (userID string, err error) {
	utils.LogHandlerF("password authorize", "username:%s password:%s", username, password)
	if username == "test" && password == "test" {
		userID = "test"
	}
	return
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	if dump {
		_ = dumpRequest(os.Stdout, "userAuthorizeHandler", r)
	}

	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		return
	}

	uid, ok := store.Get("LoggedInUserID")
	if !ok {
		if r.Form == nil {
			utils.LogHandler("userAuthorize", "r.Form is null")
			r.ParseForm()
		}

		utils.LogHandlerF("userAuthorize", "r.Form: %v", r.Form)

		store.Set("ReturnUri", r.Form)
		utils.LogHandlerF("userHandler", "set ReturnUri %v", r.Form)
		store.Save()

		redirect(w, "userAuthorize", "/login")
		utils.LogHandlerF("userAuthorize", "userid: %s, Get store.LoggedInUserID is null", userID)
		return
	}

	userID = uid.(string)
	store.Delete("LoggedInUserID")
	store.Save()
	utils.LogHandlerF("userAuthorize", "userid: %s, Delete store.LoggedInUserID", userID)
	return
}

func login(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	if dump {
		_ = dumpRequest(os.Stdout, "auth", r)
	}

	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		utils.LogRequestF("login", "method: %s url: %s", r.Method, r.URL.String())
		checkLogin(w, r, store)
		return
	}
	utils.LogRequestF("login", "method: %s url: %s", r.Method, r.URL.String())
	outputHTML(w, r, "static/login.html")
}

func checkLogin(w http.ResponseWriter, r *http.Request, store session.Store) {
	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	store.Set("LoggedInUserID", r.Form.Get("username"))
	store.Save()

	redirect(w, "login", "/auth")
}

func auth(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	utils.LogRequest("auth", r.URL)
	if dump {
		_ = dumpRequest(os.Stdout, "auth", r)
	}

	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("LoggedInUserID"); !ok {
		redirect(w, "auth", "/login")
		return
	}

	utils.LogRequestF("login", "method: %s url: %s", r.Method, r.URL)
	outputHTML(w, r, "static/auth.html")
}

func authorize(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	utils.LogRequest("oauth/authorize", r.URL)
	if dump {
		dumpRequest(os.Stdout, "authorize", r)
	}

	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
	}
	r.Form = form
	utils.LogRequestF("oauth/authorize", "Get store.ReturnUri: %v and Delete store.ReturnUri", form)

	store.Delete("ReturnUri")
	store.Save()

	err = srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func token(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	utils.LogRequest("token", r.URL)
	if dump {
		_ = dumpRequest(os.Stdout, "token", r)
	}

	err := srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func test(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	utils.LogRequest("test", r.URL)
	if dump {
		_ = dumpRequest(os.Stdout, "test", r)
	}

	token, err := srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"client_id":  token.GetClientID(),
		"user_id":    token.GetUserID(),
	}

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(data)
}

func addHandler(pattern string, handler func(w http.ResponseWriter, r *http.Request, srv *server.Server), srv *server.Server) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, srv)
	})
}

func redirect(w http.ResponseWriter, req, location string) {
	utils.LogRedirect(req, "/login")
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusFound)
}

func dumpRequest(writer io.Writer, header string, r *http.Request) error {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}

	writer.Write([]byte("\n" + header + ": \n"))
	writer.Write(data)
	return nil
}

func outputHTML(w http.ResponseWriter, r *http.Request, fileName string) {
	filePath, err := utils.GetRunPath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fullFileName := filepath.Join(filePath, fileName)
	file, err := os.Open(fullFileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, r, file.Name(), fi.ModTime(), file)
}
