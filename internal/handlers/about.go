package handlers

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/yuin/goldmark"
)

type AboutCache struct {
	sync.RWMutex
	htmlData []byte
}

var aboutCache = &AboutCache{}

func updateAboutCache(filePath string) error {
	aboutMD, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithRendererOptions(
		//html.WithUnsafe(),		//Remove comment markers in front of html.WithUnsafe() to allow HTML Tags/JavaScript/CSS Embedding in your markdown file.
		),
	)
	if err := md.Convert(aboutMD, &buf); err != nil {
		return err
	}

	aboutCache.Lock()
	aboutCache.htmlData = buf.Bytes()
	aboutCache.Unlock()

	return nil
}

func StartAboutCache() {
	filePath := os.Getenv("ABOUT_MD_PATH")
	if filePath == "" {
		filePath = "./about.md"
	}

	log.Printf("Starting initial parsing of about.md from: %s", filePath)

	if err := updateAboutCache(filePath); err == nil {
		log.Println("about.md initial cache success")
	} else {
		log.Printf("Initial about.md caching failed: %v. Trying again...", err)
	}

	ticker := time.NewTicker(60 * time.Minute) // Currently set to update the about.md every one hour, feel free to change, if you want.

	go func() {
		for range ticker.C {
			log.Println("Starting Background Caching of about.md...")
			if err := updateAboutCache(filePath); err != nil {
				log.Printf("Background Update of about.md Failed: %v", err)
				continue
			}
			log.Println("Background Update of about.md and writing to Cache success.")
		}
	}()
}

func GetHandleAbout(w http.ResponseWriter, r *http.Request) {
	aboutCache.RLock()
	html := aboutCache.htmlData
	aboutCache.RUnlock()

	if len(html) == 0 {
		http.Error(w, "About content is still loading or not available..", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}
