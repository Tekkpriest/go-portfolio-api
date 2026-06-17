package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

type CachedProjects struct {
	sync.RWMutex
	projects []GitHubProject
}

var cache = &CachedProjects{}

func fetchProjects() ([]GitHubProject, error) {
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")
	if token == "" || username == "" {
		return nil, fmt.Errorf("Github Token or Username not found in .env")
	}

	apiURL := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("User-Agent", "Go-Portfolio-API/1.0 (+https://github.com/tekkpriest/go-portfolio-api)")

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Github API Status Code: %d", response.StatusCode)
	}

	var projects []GitHubProject
	if err := json.NewDecoder(response.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func StartCacheUpdater() {
	log.Println("Starting initial download of Github projects")
	if freshProjects, err := fetchProjects(); err == nil {
		cache.Lock()
		cache.projects = freshProjects
		cache.Unlock()
		log.Println("Download success")
	} else {
		log.Printf("Download failed: %v. Trying again..", err)
	}

	ticker := time.NewTicker(10 * time.Minute)

	go func() {
		for range ticker.C {
			log.Println("Starting Background Caching...")
			freshProjects, err := fetchProjects()
			if err != nil {
				log.Printf("Background Update Failed: %v (Keeping old data)", err)
				continue
			}

			cache.Lock()
			cache.projects = freshProjects
			cache.Unlock()
			log.Println("Background Update and writing to Cache success.")
		}
	}()
}

func HandleGetProjects(w http.ResponseWriter, r *http.Request) {
	cache.RLock()
	defer cache.RUnlock()

	if cache.projects == nil {
		http.Error(w, "Projects are still loading..", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cache.projects); err != nil {
		log.Printf("JSON Encoding Error (Github Projects): %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
