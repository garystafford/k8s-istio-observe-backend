// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service A

package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	joonix "github.com/joonix/log"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
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

	err := json.NewEncoder(w).Encode(traces)
	if err != nil {
		log.WithField("func", "json.NewEncoder()").Fatal(err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.WithField("func", "w.Write()").Fatal(err)
	}
}

func CallNextService(url string) {
	var tmpTraces []Trace
	response, err := http.Get(url)
	if err != nil {
		log.WithField("func", "http.Get()").Warning(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &tmpTraces)
		if err != nil {
			log.WithField("func", "json.Unmarshal()").Warning(err)
		}

		for _, r := range tmpTraces {
			traces = append(traces, r)
		}
	}
}

func init() {
	log.SetFormatter(&joonix.FluentdFormatter{})
}

func main() {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
	})
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/ping", PingHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET", "OPTIONS")
	handler := c.Handler(router)
	err := http.ListenAndServe(":80", handler)
	if err != nil {
		log.WithField("func", "http.ListenAndServe()").Fatal(err)
	}
}
