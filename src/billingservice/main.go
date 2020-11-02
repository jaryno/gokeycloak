package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"runtime"
)

func main() {
	http.HandleFunc("/billing/v1/services", enabledLog(services))
	http.ListenAndServe(":8082", nil)
}

func services(w http.ResponseWriter, r *http.Request) {
	s := Billing{
		Services: []string{
			"electric",
			"phone",
			"internet",
			"water",
		},
	}
	// encoder := json.NewEncoder(w)
	// encoder.Encode(s)
	// w.Header().Set("Content-Type", "application/json")

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
