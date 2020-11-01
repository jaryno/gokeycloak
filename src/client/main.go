package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"learn.oatuh.client/model"
)

var port = "8081"
var host = "http://localhost:" + port

var config = struct {
	appID               string
	appPassword         string
	authURL             string
	logoutURL           string
	afterLogoutRedirect string
	authCodeCallback    string
	tokenEndpoint       string
}{
	appID:               "billingApp",
	appPassword:         "b96d3964-b8c2-4302-af07-3ec440945611",
	authURL:             "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/auth",
	logoutURL:           "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/logout",
	afterLogoutRedirect: host,
	authCodeCallback:    host + "/authCodeRedirect",
	tokenEndpoint:       "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/token",
}

type AppVar struct {
	AuthCode     string
	SessionState string
	AccessToken  string
	RefreshToken string
	Scope        string
}

var t = template.Must(template.ParseFiles("template/index.html"))
var appVar = AppVar{}

func main() {
	// fmt.Println("hello")
	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/exchangeToken", exchangeToken)
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

func exchangeToken(w http.ResponseWriter, r *http.Request) {

	// Request
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", appVar.AuthCode)
	form.Add("redirect_uri", config.authCodeCallback)
	form.Add("client_id", config.appID)
	req, err := http.NewRequest("POST", config.tokenEndpoint, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
		return
	}

	req.SetBasicAuth(config.appID, config.appPassword)

	// Client
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Println("couldn't get access token", err)
		return
	}

	// Proccess response
	byteBody, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}

	accessTokenResponse := &model.AccessTokenResponse{}
	json.Unmarshal(byteBody, accessTokenResponse)

	appVar.AccessToken = accessTokenResponse.AccessToken
	appVar.RefreshToken = accessTokenResponse.RefreshToken
	appVar.Scope = accessTokenResponse.Scope
	log.Println(appVar.AccessToken)

	t.Execute(w, appVar)
}
