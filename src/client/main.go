package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

var oauth = struct {
	authURL string
}{
	authURL: "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/auth",
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
	http.ListenAndServe(":8081", nil)
}

func home(w http.ResponseWriter, r *http.Request) {

	t.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	// create a redirect URL for authentication endpoint
	req, err := http.NewRequest("GET", oauth.authURL, nil)
	if err != nil {
		log.Print(err)
		return
	}

	qs := url.Values{}
	qs.Add("state", "123")
	qs.Add("client_id", "billingApp")
	qs.Add("response_type", "code")
	qs.Add("redirect_uri", "http://localhost:8081/authCodeRedirect")

	// "state=123abc&client_id=billingApp&response_type=code"
	req.URL.RawQuery = qs.Encode()

	http.Redirect(w, r, req.URL.String(), http.StatusFound)
}

func authCodeRedirect(w http.ResponseWriter, r *http.Request) {
	appVar.AuthCode = r.URL.Query().Get("code")
	appVar.SessionState = r.URL.Query().Get("session_state")
	r.URL.RawQuery = ""
	fmt.Printf("Request queries: %+v\n", appVar)

	http.Redirect(w, r, "http://localhost:8081/", http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {

}
