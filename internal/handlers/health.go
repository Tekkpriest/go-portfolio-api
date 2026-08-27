package handlers

import (
	"net/http"
	"time"
)

type HealthChecker interface {
	LastSuccessfulRefresh() time.Time
}

type HealthHandler struct {
	about    HealthChecker
	projects HealthChecker
	maxAge   time.Duration
}

func NewHealthHandler(about, projects HealthChecker, maxAge time.Duration) *HealthHandler {
	return &HealthHandler{about: about, projects: projects, maxAge: maxAge}
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	aboutAge := now.Sub(h.about.LastSuccessfulRefresh())
	projectsAge := now.Sub(h.projects.LastSuccessfulRefresh())

	healthy := aboutAge < h.maxAge && projectsAge < h.maxAge

	status := "ok"
	httpStatus := http.StatusOK

	if !healthy {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	WriteJSONResponse(w, map[string]any{
		"status":         status,
		"about_age_s":    int(aboutAge.Seconds()),
		"projects_age_s": int(projectsAge.Seconds()),
	}, httpStatus)
}
