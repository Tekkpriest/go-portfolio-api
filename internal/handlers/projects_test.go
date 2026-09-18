package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/tekkpriest/go-portfolio-api/internal/caching"
)

type mockProjectSource struct {
	projects []caching.GitHubProject
	err      error
}

func (m *mockProjectSource) GetProjects() ([]caching.GitHubProject, error) {
	return m.projects, m.err
}

func TestProjectHandler_GetProjects(t *testing.T) {
	sampleProject := caching.GitHubProject{
		ID:          1,
		Name:        "go-test-portfolio",
		Description: "repo for testing purposes",
		Language:    "Go",
		Tags:        []string{"backend", "portfolio", "html"},
		URL:         "https://github.com/testuser/go-test-portfolio",
		Stars:       5,
	}

	tests := []struct {
		TestName       string
		SourceProjects []caching.GitHubProject
		SourceErr      error
		WantStatusCode int
		WantErrMessage string
	}{
		{
			TestName:       "returns projects on success",
			SourceProjects: []caching.GitHubProject{sampleProject},
			WantStatusCode: http.StatusOK,
		},
		{
			TestName:       "returns empty but valid list",
			SourceProjects: []caching.GitHubProject{},
			WantStatusCode: http.StatusOK,
		},
		{
			TestName:       "returns 503 when source errors",
			SourceErr:      errors.New("cache empty"),
			WantStatusCode: http.StatusServiceUnavailable,
			WantErrMessage: "projects are still loading",
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockSource := &mockProjectSource{projects: tc.SourceProjects, err: tc.SourceErr}
			handler := NewProjectHandler(mockSource)

			req := httptest.NewRequestWithContext(t.Context(), "GET", "/api/projects", nil)
			rec := httptest.NewRecorder()

			handler.GetProjects(rec, req)

			if rec.Code != tc.WantStatusCode {
				t.Errorf("expected status %d, got %d", tc.WantStatusCode, rec.Code)
			}

			if tc.SourceErr != nil {
				var errResp APIError
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response body: %v", err)
				}
				if errResp.Error != tc.WantErrMessage {
					t.Errorf("expected error message %q, got %q", tc.WantErrMessage, errResp.Error)
				}
				return
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", ct)
			}

			var got []caching.GitHubProject
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			if !reflect.DeepEqual(got, tc.SourceProjects) {
				t.Errorf("expected projects %+v, got %+v", tc.SourceProjects, got)
			}
		})
	}
}
