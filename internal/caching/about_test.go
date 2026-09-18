package caching

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAboutCache_GetHTML(t *testing.T) {
	tests := []struct {
		TestName   string
		SetupCache func(t *testing.T) *AboutCache
		WantErr    bool
		WantHTML   string
	}{
		{
			TestName: "returns rendered HTML after successful refresh",
			SetupCache: func(t *testing.T) *AboutCache {
				t.Helper()
				dir := t.TempDir()
				filePath := filepath.Join(dir, "about.md")
				if err := os.WriteFile(filePath, []byte("# Hello World"), 0644); err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				cache := NewAboutCache(filePath)
				if err := cache.refresh(); err != nil {
					t.Fatalf("failed to refresh cache: %v", err)
				}
				return cache
			},
			WantHTML: "<h1>Hello World</h1>\n",
		},
		{
			TestName: "returns error when no refresh has happened yet",
			SetupCache: func(t *testing.T) *AboutCache {
				t.Helper()
				return NewAboutCache("irrelevant.md")
			},
			WantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			cache := tc.SetupCache(t)
			html, err := cache.GetHTML()

			if tc.WantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected html content, got error: %v", err)
			}
			if string(html) != tc.WantHTML {
				t.Errorf("expected html %q, got %q", tc.WantHTML, string(html))
			}
		})
	}
}

func TestAboutCache_Start_RefreshesInBackground(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "about.md")
	if err := os.WriteFile(filePath, []byte("# Hello World"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cache := NewAboutCache(filePath)

	cache.Start(t.Context())

	const wantHTML = "<h1>Hello World</h1>\n"
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if html, err := cache.GetHTML(); err == nil {
			if string(html) != wantHTML {
				t.Errorf("expected html %q, got %q", wantHTML, string(html))
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("timed out waiting for background refresh to populate the cache")
}

func TestAboutCache_Refresh_MissingFile(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "does-not-exist.md")
	cache := NewAboutCache(missingPath)

	if err := cache.refresh(); err == nil {
		t.Error("expected refresh to fail for a missing file")
	}
}
