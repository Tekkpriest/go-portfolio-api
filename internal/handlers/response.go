package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type APIError struct {
	Error string `json:"error"`
}

func WriteJSONResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed json encoding", "error", err)
	}
}

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, clientMsg string, internalError error, status int) {
	if status >= 500 {
		slog.Error("http request failed",
			"status", status,
			"path", r.URL.Path,
			"method", r.Method,
			"error", internalError)
	} else {
		slog.Warn("http client error",
			"status", status,
			"path", r.URL.Path,
			"method", r.Method,
			"msg", clientMsg)
	}

	WriteJSONResponse(w, APIError{Error: clientMsg}, status)
}
