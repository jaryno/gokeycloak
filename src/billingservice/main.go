package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

type Billing struct {
	Services []string `json:"services"`
}

type BillingError struct {
	Error string `json:"error"`
}

// Introspect response
type TokenIntrospect struct {
	Jti      string `json:"jti"`
	Exp      int    `json:"exp"`
	Nbf      int    `json:"nbf"`
	Iat      int    `json:"iat"`
	Aud      string `json:"aud"`
	Typ      string `json:"typ"`
	AuthTime int    `json:"auth_time"`
	Acr      string `json:"acr"`
	Active   bool   `json:"active"`
}

var config = struct {
	tokenIntroSpection string
}{
	tokenIntroSpection: "http://localhost:8080/auth/realms/learningApp/protocol/openid-connect/token/introspect",
}

func main() {
	http.HandleFunc("/billing/v1/services", enabledLog(services))
	http.ListenAndServe(":8082", nil)
}

func services(w http.ResponseWriter, r *http.Request) {

	token, err := getToken(r)
	if err != nil {
		log.Println(err)
		s := &BillingError{Error: err.Error()}
		encoder := json.NewEncoder(w)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(s)
		return
	}

	log.Println("Token : ", token)

	// valdate token
	if !validateToken(token) {
		log.Println(err)
		s := &BillingError{Error: "InvalidToken"}
		encoder := json.NewEncoder(w)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(s)
		return
	}

	s := Billing{
		Services: []string{
			"electric",
			"phone",
			"internet",
			"water",
		},
	}
	// encoder := json.NewEncoder(w)
	// w.Header().Add("Content-Type", "application/json")
	// encoder.Encode(s)

	jData, err := json.Marshal(s)
	if err != nil {
		log.Print(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func enabledLog(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		log.SetPrefix(handlerName + " ")
		log.Println("--> " + handlerName)
		log.Printf("request : %+v\n", r.RequestURI)
		// log.Printf("response : %+v\n", w)
		handler(w, r)
		log.Println("<-- " + handlerName)
	}
}

func getToken(r *http.Request) (string, error) {
	// header
	token := r.Header.Get("Authorization")
	if token != "" {
		auths := strings.Split(token, " ")
		if len(auths) != 2 {
			return "", fmt.Errorf("invalid Authorization header format")
		}
		return auths[1], nil
	}

	// form body
	token = r.FormValue("access_token")
	if token != "" {
		return token, nil
	}

	// query
	token = r.URL.Query().Get("access_token")
	if token != "" {
		return token, nil
	}

	return token, fmt.Errorf("Missing access token")
}

func validateToken(token string) bool {

	// request
	form := url.Values{}
	form.Add("token", token)
	form.Add("token_type_hint", "requesting_party_token")
	req, err := http.NewRequest("POST", config.tokenIntroSpection, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println(err)
		return false
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("tokenChecker", "0c83b0b9-00ea-4ac9-a7d4-0a67f9cd10a3")

	// client
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}

	// process response
	if res.StatusCode != 200 {
		log.Println("Status is not 200 : ", res.StatusCode)
		return false
	}

	byteBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	defer res.Body.Close()

	introSpect := &TokenIntrospect{}
	err = json.Unmarshal(byteBody, introSpect)
	if err != nil {
		log.Println(err)
		return false
	}

	return introSpect.Active
}
