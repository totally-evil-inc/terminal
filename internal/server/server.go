package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Application struct {
	server *http.Server
	config *Config
}

type Config struct {
	Port            int
	ShutdownTimeout time.Duration
	// DbUrl string
}

func NewApplication(cfg *Config) *Application {
	return &Application{
		server: NewServer(cfg),
		config: cfg,
	}
}

func NewConfig() *Config {
	return &Config{
		Port:            8080,
		ShutdownTimeout: time.Duration(time.Second * 5),
	}
}

func NewServer(cfg *Config) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", testRoute)
	
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return srv
}

func (a *Application) Run() error {
	return a.server.ListenAndServe()
}

func testRoute(w http.ResponseWriter, r *http.Request) {
	_, err := r.Body.Read([]byte{})
	if err != nil {
		log.Printf("Error occurred %v\n", err.Error())
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode("Naah we good")
}