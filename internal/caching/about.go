package caching

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/yuin/goldmark"
)

type AboutCache struct {
	mu                   sync.RWMutex
	htmlData             []byte
	filePath             string
	lastSuccessfulUpdate time.Time
}

func NewAboutCache(filePath string) *AboutCache {
	return &AboutCache{filePath: filePath}
}

func (cache *AboutCache) Start(ctx context.Context) {
	go func() {
		if err := cache.refresh(); err != nil {
			slog.Error("initial about.md caching failed", "error", err, "path", cache.filePath)
		} else {
			slog.Info("initial about.md caching successful", "path", cache.filePath)
		}

		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("stopping about.md background update")
				return
			case <-ticker.C:
				slog.Info("starting background update of about.md")
				if err := cache.refresh(); err != nil {
					slog.Error("background update of about.md failed", "error", err)
					continue
				}
				slog.Info("background update of about.md successful")
			}
		}
	}()
}

func (cache *AboutCache) GetHTML() ([]byte, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	if cache.htmlData == nil {
		return nil, fmt.Errorf("about.md content not yet available")
	}

	data := make([]byte, len(cache.htmlData))
	copy(data, cache.htmlData)

	return data, nil
}

func (cache *AboutCache) refresh() error {
	aboutMD, err := os.ReadFile(cache.filePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithRendererOptions(
		// html.WithUnsafe(),  << okay to use if only you have access to the markdown files, otherwise remove.
		),
	)

	if err := md.Convert(aboutMD, &buf); err != nil {
		return err
	}

	cache.mu.Lock()
	cache.htmlData = buf.Bytes()
	cache.lastSuccessfulUpdate = time.Now()
	cache.mu.Unlock()

	return nil
}

func (cache *AboutCache) LastSuccessfulRefresh() time.Time {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	return cache.lastSuccessfulUpdate
}
