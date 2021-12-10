package schema

import (
	"encoding/json"
	"log"
	"net/http"
)

type LambdaHandlerError struct {
	RequestID  string `json:"request_id,omitempty"`
	ErrorType  string `json:"error_type,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (e *LambdaHandlerError) String() string {
	e.ErrorType = http.StatusText(e.StatusCode)
	b, err := json.Marshal(e)
	if err != nil {
		log.Printf("error marshalling lambda error: %s\n", err)
		return "error building response"
	}
	return string(b)
}
