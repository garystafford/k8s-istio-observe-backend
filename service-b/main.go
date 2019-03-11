// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service B

package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Trace struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"service,omitempty"`
	Greeting    string    `json:"greeting,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
}

var traces []Trace

func PingHandler(w http.ResponseWriter, r *http.Request) {
	traces = nil
	CallNextService("http://service-d/api/ping")
	CallNextService("http://service-e/api/ping")

	tmpTrace := Trace{
		ID:          uuid.New().String(),
		ServiceName: "Service-B",
		Greeting:    "Namaste, from Service-B!",
		CreatedAt:   time.Now().Local(),
	}

	traces = append(traces, tmpTrace)
	fmt.Println(traces)

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		log.Fatal(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Fatal(err)
	}
}

func CallNextService(url string) {
	var tmpTraces []Trace
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &tmpTraces)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range tmpTraces {
			traces = append(traces, r)
		}
	}
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}
