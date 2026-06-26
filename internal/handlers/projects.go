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

var projectCache = &CachedProjects{}

func fetchProjects() ([]GitHubProject, error) {
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")
	if token == "" || username == "" {
		return nil, fmt.Errorf("Github Token or Username not in .env")
	}

	apiURL := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("User-Agent", "Portfolio/1.0 (+https://github.com/tekkpriest/go-portfolio-api)")
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

func StartProjectCache() {
	log.Println("Starting initial download of Github projects")
	if freshProjects, err := fetchProjects(); err == nil {
		projectCache.Lock()
		projectCache.projects = freshProjects
		projectCache.Unlock()
		log.Println("Download of projects successful")
	} else {
		log.Printf("Downloading projects failed: %v. Trying again..", err)
	}

	ticker := time.NewTicker(20 * time.Minute) // Currently set to update GH projects every 20 Minutes, feel free to change.

	go func() {
		for range ticker.C {
			log.Println("Starting Background Caching of Projects...")
			freshProjects, err := fetchProjects()
			if err != nil {
				log.Printf("Background Update of Projects Failed: %v", err)
				continue
			}

			projectCache.Lock()
			projectCache.projects = freshProjects
			projectCache.Unlock()
			log.Println("Background Update of Projects and writing to Cache success.")
		}
	}()
}

func GetHandleProjects(w http.ResponseWriter, r *http.Request) {
	projectCache.RLock()
	defer projectCache.RUnlock()

	if projectCache.projects == nil {
		http.Error(w, "Projects are still loading..", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projectCache.projects); err != nil {
		log.Printf("JSON Encoding Error (Github Projects): %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
