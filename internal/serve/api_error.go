package serve

import (
	"encoding/json"
	"net/http"
)

type apiErrorJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeAPIErrorJSON(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiErrorJSON{Code: code, Message: message})
}
