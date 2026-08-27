package handlers

import (
	"net/http"
)

type AboutSource interface {
	GetHTML() ([]byte, error)
}

type AboutHandler struct {
	aboutCache AboutSource
}

func NewAboutHandler(a AboutSource) *AboutHandler {
	return &AboutHandler{aboutCache: a}
}

func (h *AboutHandler) GetAbout(w http.ResponseWriter, r *http.Request) {
	body, err := h.aboutCache.GetHTML()
	if err != nil {
		WriteErrorResponse(w, r, "about is still loading or not available", err, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(body); err != nil {
		WriteErrorResponse(w, r, "failed to write response", err, http.StatusInternalServerError)
	}
}
