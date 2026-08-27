package handlers

import (
	"net/http"

	"github.com/tekkpriest/go-portfolio-api/internal/caching"
)

type ProjectSource interface {
	GetProjects() ([]caching.GitHubProject, error)
}

type ProjectHandler struct {
	projectCache ProjectSource
}

func NewProjectHandler(p ProjectSource) *ProjectHandler {
	return &ProjectHandler{projectCache: p}
}

func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectCache.GetProjects()
	if err != nil {
		WriteErrorResponse(w, r, "projects are still loading", err, http.StatusServiceUnavailable)
		return
	}

	WriteJSONResponse(w, projects, http.StatusOK)
}
