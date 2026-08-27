package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type GitHubProject struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Tags        []string `json:"topics"`
	URL         string   `json:"html_url"`
	Stars       int      `json:"stargazers_count"`
}

type ProjectCache struct {
	mu                   sync.RWMutex
	projects             []GitHubProject
	token                string
	username             string
	apiURL               string
	lastSuccessfulUpdate time.Time
	httpClient           *http.Client
}

func NewProjectCache(token, username string) *ProjectCache {
	return &ProjectCache{
		token:      token,
		username:   username,
		apiURL:     "https://api.github.com",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (cache *ProjectCache) Start(ctx context.Context) {
	go func() {
		if err := cache.refresh(ctx); err != nil {
			slog.Error("initial github projects fetch failed", "error", err)
		} else {
			slog.Info("initial github projects cache successful")
		}

		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("stopping github project background updates")
				return
			case <-ticker.C:
				slog.Info("starting github project background update")
				if err := cache.refresh(ctx); err != nil {
					slog.Error("github background update failed", "error", err)
					continue
				}
				slog.Info("github background update success")
			}
		}
	}()
}

func (cache *ProjectCache) GetProjects() ([]GitHubProject, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if cache.projects == nil {
		return nil, fmt.Errorf("projects cache is empty")
	}

	data := make([]GitHubProject, len(cache.projects))
	copy(data, cache.projects)

	return data, nil
}

func (cache *ProjectCache) refresh(ctx context.Context) error {
	apiURL := fmt.Sprintf("%s/users/%s/repos", cache.apiURL, cache.username)
	request, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+cache.token)
	request.Header.Set("User-Agent", "Portfolio/1.0 (+https://github.com/tekkpriest/go-portfolio-api)")

	response, err := cache.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("github api status: %d", response.StatusCode)
	}

	var newProjects []GitHubProject
	if err := json.NewDecoder(response.Body).Decode(&newProjects); err != nil {
		return err
	}

	cache.mu.Lock()
	cache.projects = newProjects
	cache.lastSuccessfulUpdate = time.Now()
	cache.mu.Unlock()

	return nil
}

func (cache *ProjectCache) LastSuccessfulRefresh() time.Time {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	return cache.lastSuccessfulUpdate
}
