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
	flag.StringVar(&secret, "s", "333333", "The client secret being passed in")
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
	log.Printf("Point your OAuth client Token endpoint to %s%s", endpoint, "")

	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func passwordAuthorizeHandler(ctx context.Context, clientID, username, password string) (userID string, err error) {
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
			r.ParseForm()
		}

		store.Set("ReturnUri", r.Form)
		store.Save()

		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	userID = uid.(string)
	store.Delete("LoggedInUserID")
	store.Save()
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
		checkLogin(w, r, store)
		return
	}
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

	w.Header().Set("Location", "/auth")
	w.WriteHeader(http.StatusFound)
}

func auth(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	if dump {
		_ = dumpRequest(os.Stdout, "auth", r)
	}

	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	outputHTML(w, r, "static/auth.html")
}

func authorize(w http.ResponseWriter, r *http.Request, srv *server.Server) {
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

	store.Delete("ReturnUri")
	store.Save()

	err = srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func token(w http.ResponseWriter, r *http.Request, srv *server.Server) {
	if dump {
		_ = dumpRequest(os.Stdout, "token", r)
	}

	err := srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func test(w http.ResponseWriter, r *http.Request, srv *server.Server) {
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
	file, err := os.Open(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, r, file.Name(), fi.ModTime(), file)
}
