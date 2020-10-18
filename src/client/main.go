package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

var port = "8081"
var host = "http://localhost:" + port

var config = struct {
	appID               string
	authURL             string
	logoutURL           string
	afterLogoutRedirect string
	authCodeCallback    string
}{
	appID:               "billingApp",
	authURL:             "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/auth",
	logoutURL:           "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/logout",
	afterLogoutRedirect: host,
	authCodeCallback:    host + "/authCodeRedirect",
}

type AppVar struct {
	AuthCode     string
	SessionState string
}

var t = template.Must(template.ParseFiles("template/index.html"))
var appVar = AppVar{}

func main() {
	// fmt.Println("hello")
	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/authCodeRedirect", authCodeRedirect)
	http.ListenAndServe(":"+port, nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	t.Execute(w, appVar)
}

func login(w http.ResponseWriter, r *http.Request) {
	// create a redirect URL for authentication endpoint
	req, err := http.NewRequest("GET", config.authURL, nil)
	if err != nil {
		log.Print(err)
		return
	}

	qs := url.Values{}
	qs.Add("state", "123")
	qs.Add("client_id", config.appID)
	qs.Add("response_type", "code")
	qs.Add("redirect_uri", config.authCodeCallback)

	req.URL.RawQuery = qs.Encode()

	http.Redirect(w, r, req.URL.String(), http.StatusFound)
}

func authCodeRedirect(w http.ResponseWriter, r *http.Request) {
	appVar.AuthCode = r.URL.Query().Get("code")
	appVar.SessionState = r.URL.Query().Get("session_state")
	r.URL.RawQuery = ""
	fmt.Printf("Request queries: %+v\n", appVar)

	http.Redirect(w, r, host, http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	q := url.Values{}
	q.Add("redirect_uri", config.afterLogoutRedirect)

	logoutURL, err := url.Parse(config.logoutURL)
	if err != nil {
		log.Println(err)
	}
	logoutURL.RawQuery = q.Encode()
	appVar = AppVar{}
	http.Redirect(w, r, logoutURL.String(), http.StatusFound)
}
