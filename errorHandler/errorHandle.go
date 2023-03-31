package errorHandler

import (
	"encoding/json"
	"net/http"
)

type CustomError struct {
	Message string `json:"message"`
}

func NewHandlerError(w http.ResponseWriter, message string, errorCode int) {
	w.WriteHeader(errorCode)
	customError := CustomError{
		Message: message,
	}
	json.NewEncoder(w).Encode(customError)
}
