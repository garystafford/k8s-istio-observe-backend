// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Service A
// date: 2021-06-05

package main

import (
	"encoding/json"
	"fmt"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var (
	logLevel       = getEnv("LOG_LEVEL", "debug")
	port           = getEnv("PORT", ":8080")
	message        = getEnv("GREETING", "Hello, from Service A!")
	allowedOrigins = getEnv("ALLOWED_ORIGINS", "*")
	URLServiceB    = getEnv("SERVICE_B_URL", "http://service-b")
	URLServiceC    = getEnv("SERVICE_C_URL", "http://service-c")
)

type Greeting struct {
	ID          string    `json:"id,omitempty"`
	ServiceName string    `json:"service,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created,omitempty"`
	Hostname    string    `json:"hostname,omitempty"`
}

var greetings []Greeting

func GreetingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	log.Debug(r)

	greetings = nil

	callNextServiceWithTrace(URLServiceB+"/api/greeting", w, r)
	callNextServiceWithTrace(URLServiceC+"/api/greeting", w, r)

	tmpGreeting := Greeting{
		ID:          uuid.New().String(),
		ServiceName: "Service A",
		Message:     message,
		CreatedAt:   time.Now().Local(),
		Hostname:    getHostname(),
	}

	greetings = append(greetings, tmpGreeting)

	err := json.NewEncoder(w).Encode(greetings)
	if err != nil {
		log.Error(err)
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
	}
	return hostname
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"alive\": true}"))
	if err != nil {
		log.Error(err)
	}
}

func ResponseStatusHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	statusCode, err := strconv.Atoi(params["code"])
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
}

func callNextServiceWithTrace(url string, w http.ResponseWriter, r *http.Request) {
	var tmpGreetings []Greeting

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
	}

	// Headers must be passed for Jaeger Distributed Tracing
	headers := []string{
		"uber-trace-id",
		"x-b3-flags",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-spanid",
		"x-b3-traceid",
		"x-ot-span-context",
		"x-request-id",
	}

	for _, header := range headers {
		if r.Header.Get(header) != "" {
			req.Header.Add(header, r.Header.Get(header))
		}
	}

	log.Info(req)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Do(req)

	if err != nil {
		log.Error(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(response.Body)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
	}

	err = json.Unmarshal(body, &tmpGreetings)
	if err != nil {
		log.Error(err)
	}

	for _, r := range tmpGreetings {
		greetings = append(greetings, r)
	}
}

func RequestEchoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Error(err)
	}
	_, err = fmt.Fprintf(w, string(requestDump))
	if err != nil {
		log.Error(err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func run() error {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigins},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
	})
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/greeting", GreetingHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/health", HealthCheckHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/request-echo", RequestEchoHandler).Methods(
		"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD")
	api.HandleFunc("/status/{code}", ResponseStatusHandler).Methods("GET", "OPTIONS")
	api.Handle("/metrics", promhttp.Handler())
	handler := c.Handler(router)
	return http.ListenAndServe(port, handler)
}

func init() {
	formatter := runtime.Formatter{ChildFormatter: &log.JSONFormatter{}}
	formatter.Line = true
	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Error(err)
	}
	log.SetLevel(level)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
