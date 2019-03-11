// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service A

package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	traces = nil
	CallNextService("http://service-b/api/ping")
	CallNextService("http://service-c/api/ping")

	tmpTrace := Trace{
		ID:          uuid.New().String(),
		ServiceName: "Service-A",
		Greeting:    "Hello, from Service-A!",
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
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET", "OPTIONS")
	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":80", handler))
}
