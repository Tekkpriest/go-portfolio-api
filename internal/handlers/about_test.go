package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAboutSource struct {
	html []byte
	err  error
}

func (m *mockAboutSource) GetHTML() ([]byte, error) {
	return m.html, m.err
}

func TestAboutHandler_GetAbout(t *testing.T) {
	tests := []struct {
		TestName       string
		SourceHTML     []byte
		SourceErr      error
		WantStatusCode int
		WantBody       string
		WantErrMessage string
	}{
		{
			TestName:       "returns rendered HTML on success",
			SourceHTML:     []byte("<p>Hello</p>"),
			WantStatusCode: http.StatusOK,
			WantBody:       "<p>Hello</p>",
		},
		{
			TestName:       "returns 503 when source errors",
			SourceErr:      errors.New("not cached yet"),
			WantStatusCode: http.StatusServiceUnavailable,
			WantErrMessage: "about is still loading or not available",
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockSource := &mockAboutSource{html: tc.SourceHTML, err: tc.SourceErr}
			handler := NewAboutHandler(mockSource)

			req := httptest.NewRequestWithContext(t.Context(), "GET", "/api/aboutme", nil)
			rec := httptest.NewRecorder()

			handler.GetAbout(rec, req)

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

			if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Errorf("expected Content-Type text/html, got %q", ct)
			}
			if rec.Body.String() != tc.WantBody {
				t.Errorf("expected body %q, got %q", tc.WantBody, rec.Body.String())
			}
		})
	}
}
