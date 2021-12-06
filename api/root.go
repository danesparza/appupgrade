package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// Service encapsulates API service operations
type Service struct {
	StartTime time.Time
}

// SystemResponse is a response for a system request
type SystemResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an API response
type ErrorResponse struct {
	Message string `json:"message"`
}

// sendErrorResponse is used to send back an error:
func sendErrorResponse(rw http.ResponseWriter, err error, code int) {
	//	Our return value
	response := ErrorResponse{Message: "Error: " + err.Error()}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(response)
}
