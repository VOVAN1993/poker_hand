package server

import (
	"encoding/json"
	"net/http"
)

func ServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("Server encountered an error."))
}

func RespondError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(message))
}
func RespondJSON(w http.ResponseWriter, status int, data interface{}) error {
	response, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write(response)

	return nil
}
