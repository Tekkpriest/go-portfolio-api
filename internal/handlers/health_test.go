package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockHealthChecker struct {
	lastRefresh time.Time
}

func (m *mockHealthChecker) LastSuccessfulRefresh() time.Time {
	return m.lastRefresh
}

func TestHealthHandler_GetHealth(t *testing.T) {
	fresh := time.Now()
	stale := time.Now().Add(-2 * time.Hour)

	tests := []struct {
		TestName            string
		AboutLastRefresh    time.Time
		ProjectsLastRefresh time.Time
		WantStatusCode      int
	}{
		{TestName: "both fresh", AboutLastRefresh: fresh, ProjectsLastRefresh: fresh, WantStatusCode: http.StatusOK},
		{TestName: "only about is stale", AboutLastRefresh: stale, ProjectsLastRefresh: fresh, WantStatusCode: http.StatusServiceUnavailable},
		{TestName: "only projects is stale", AboutLastRefresh: fresh, ProjectsLastRefresh: stale, WantStatusCode: http.StatusServiceUnavailable},
		{TestName: "both stale", AboutLastRefresh: stale, ProjectsLastRefresh: stale, WantStatusCode: http.StatusServiceUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			about := &mockHealthChecker{lastRefresh: tc.AboutLastRefresh}
			projects := &mockHealthChecker{lastRefresh: tc.ProjectsLastRefresh}
			handler := NewHealthHandler(about, projects, 90*time.Minute)

			req := httptest.NewRequestWithContext(t.Context(), "GET", "/api/health", nil)
			rec := httptest.NewRecorder()
			handler.GetHealth(rec, req)

			if rec.Code != tc.WantStatusCode {
				t.Errorf("expected status %d, got %d", tc.WantStatusCode, rec.Code)
			}
		})
	}
}
