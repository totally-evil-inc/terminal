package router

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type TestHandler struct{}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

type testResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
	Value     int    `json:"value"`
}

func (h *TestHandler) test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := testResponse{
		Message:   randomMessage(),
		RequestID: middleware.GetReqID(r.Context()),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Value:     rand.Intn(1000),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode test response: %v", err)
	}
}

func randomMessage() string {
	messages := []string{
		"test endpoint is working",
		"random response generated",
		"request received",
	}

	return messages[rand.Intn(len(messages))]
}
