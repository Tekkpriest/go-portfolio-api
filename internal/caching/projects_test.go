package caching

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestProjectCache_Refresh_Success(t *testing.T) {
	t.Parallel()

	mockGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/users/testuser/repos" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"id": 101, "name": "go-test-portfolio", "stargazers_count": 42, "language": "Go"},
			{"id": 102, "name": "go-test-webcrawler", "stargazers_count": 10, "language": "Go"}
		]`))
	}))
	defer mockGitHub.Close()

	cache := NewProjectCache("test-token", "testuser")
	cache.apiURL = mockGitHub.URL
	cache.httpClient = mockGitHub.Client()

	if err := cache.refresh(t.Context()); err != nil {
		t.Fatalf("expected refresh to succeed, got: %v", err)
	}

	projects, err := cache.GetProjects()
	if err != nil {
		t.Fatalf("expected GetProjects to return data, got error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects in cache, got %d", len(projects))
	}
	if projects[0].Name != "go-test-portfolio" || projects[0].Stars != 42 {
		t.Errorf("unexpected content in first project: %+v", projects[0])
	}
}

func TestProjectCache_Refresh_Errors(t *testing.T) {
	tests := []struct {
		TestName       string
		WantStatusCode int
		WantBody       string
	}{
		{TestName: "GitHub API responds with error status", WantStatusCode: http.StatusInternalServerError, WantBody: ""},
		{TestName: "response is not valid JSON", WantStatusCode: http.StatusOK, WantBody: "{not-valid-json"},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.WantStatusCode)
				w.Write([]byte(tc.WantBody))
			}))
			defer mockGitHub.Close()

			cache := NewProjectCache("test-token", "testuser")
			cache.apiURL = mockGitHub.URL
			cache.httpClient = mockGitHub.Client()
			cache.projects = []GitHubProject{{ID: 1, Name: "old-test-cache"}}

			if err := cache.refresh(t.Context()); err == nil {
				t.Fatal("expected refresh to fail")
			}

			projects, _ := cache.GetProjects()
			if len(projects) != 1 || projects[0].Name != "old-test-cache" {
				t.Errorf("failed refresh should not touch the old cache, got %+v", projects)
			}
		})
	}
}

func TestProjectCache_ThreadSafety(t *testing.T) {
	t.Parallel()

	cache := NewProjectCache("test-token", "testuser")
	cache.projects = []GitHubProject{{ID: 1, Name: "firstTest"}}

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := cache.GetProjects(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			cache.mu.Lock()
			cache.projects = []GitHubProject{{ID: 2, Name: "updatedTest"}}
			cache.mu.Unlock()
		}()
	}
	wg.Wait()
}
