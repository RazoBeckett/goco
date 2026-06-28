package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// modelsdev.go — models.dev registry integration for model listing.
//
// Fetches from https://models.dev/api.json — a community-maintained database
// of 4000+ models across 109+ providers. Cache strategy mirrors hermes-agent:
// in-memory (1h TTL) → disk cache → network → stale disk fallback.
//
// This lets `goco models` work without an API key and makes model listing
// near-instant on repeated runs.

const modelsDevURL = "https://models.dev/api.json"
const modelsDevCacheTTL = 1 * time.Hour
const modelsDevStaleGrace = 5 * time.Minute

// providerToModelsDev maps GoCo provider names to models.dev provider IDs.
var providerToModelsDev = map[string]string{
	ProviderGemini: "google",
	ProviderGroq:   "groq",
}

// Patterns for non-agentic / noise models to exclude.
var modelNoisePattern = regexp.MustCompile(
	`(?i)-tts\b|embedding|live-|-(preview|exp)-\d{2,4}[-_]|-image\b|-image-preview\b|-customtools\b`,
)

var (
	modelsDevMu    sync.RWMutex
	modelsDevCache map[string]json.RawMessage
	modelsDevTime  time.Time
)

// FetchModelsDev returns the full models.dev provider registry.
// Cache hierarchy: in-memory → disk → network → stale disk fallback.
func FetchModelsDev() (map[string]json.RawMessage, error) {
	modelsDevMu.RLock()
	if modelsDevCache != nil && time.Since(modelsDevTime) < modelsDevCacheTTL {
		data := modelsDevCache
		modelsDevMu.RUnlock()
		return data, nil
	}
	modelsDevMu.RUnlock()

	modelsDevMu.Lock()
	defer modelsDevMu.Unlock()

	// Double-check after acquiring write lock.
	if modelsDevCache != nil && time.Since(modelsDevTime) < modelsDevCacheTTL {
		return modelsDevCache, nil
	}

	// Try disk cache.
	if data := loadModelsDevDiskCache(); data != nil {
		modelsDevCache = data
		modelsDevTime = diskCacheMTime()
		if modelsDevTime.IsZero() {
			modelsDevTime = time.Now()
		}
		return data, nil
	}

	// Network fetch.
	data, err := fetchModelsDevNetwork()
	if err == nil && data != nil {
		modelsDevCache = data
		modelsDevTime = time.Now()
		saveModelsDevDiskCache(data)
		return data, nil
	}

	// Network failed — try stale disk cache.
	if data := loadModelsDevDiskCache(); data != nil {
		modelsDevCache = data
		modelsDevTime = time.Now().Add(-modelsDevCacheTTL + modelsDevStaleGrace)
		return data, nil
	}

	return nil, fmt.Errorf("models.dev registry unavailable: %w", err)
}

// ListModelsFromDev returns model IDs for a GoCo provider from models.dev.
// Returns nil if the provider is unknown or models.dev is unreachable.
func ListModelsFromDev(providerName string) ([]string, error) {
	mdevID, ok := providerToModelsDev[providerName]
	if !ok {
		return nil, nil
	}

	data, err := FetchModelsDev()
	if err != nil {
		return nil, err
	}

	raw, ok := data[mdevID]
	if !ok {
		return nil, nil
	}

	var providerData struct {
		Models map[string]json.RawMessage `json:"models"`
	}
	if err := json.Unmarshal(raw, &providerData); err != nil {
		return nil, fmt.Errorf("parse models.dev data for %s: %w", mdevID, err)
	}

	var models []string
	for modelID, modelRaw := range providerData.Models {
		if shouldHideModel(providerName, modelID) {
			continue
		}

		var entry struct {
			ToolCall bool `json:"tool_call"`
		}
		_ = json.Unmarshal(modelRaw, &entry)

		if !entry.ToolCall {
			continue
		}
		if modelNoisePattern.MatchString(modelID) {
			continue
		}

		models = append(models, modelID)
	}

	return models, nil
}

// shouldHideModel filters out known problematic models.
func shouldHideModel(provider, modelID string) bool {
	lower := strings.ToLower(modelID)
	if provider == ProviderGemini {
		// Hide Gemma models (low TPM quotas) and stale/retired Gemini slugs.
		switch lower {
		case "gemma-4-31b-it", "gemma-4-26b-it", "gemma-4-26b-a4b-it",
			"gemma-3-1b", "gemma-3-1b-it", "gemma-3-2b", "gemma-3-2b-it",
			"gemma-3-4b", "gemma-3-4b-it", "gemma-3-12b", "gemma-3-12b-it",
			"gemma-3-27b", "gemma-3-27b-it",
			"gemini-1.5-flash", "gemini-1.5-pro", "gemini-1.5-flash-8b",
			"gemini-2.0-flash", "gemini-2.0-flash-lite":
			return true
		}
	}
	return false
}

// --- Disk cache ---

func modelsDevCachePath() string {
	configDir := os.Getenv("XDG_CACHE_HOME")
	if configDir == "" {
		homeDir, _ := os.UserHomeDir()
		if homeDir == "" {
			return ""
		}
		configDir = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(configDir, "goco", "models-dev-cache.json")
}

func loadModelsDevDiskCache() map[string]json.RawMessage {
	path := modelsDevCachePath()
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var data map[string]json.RawMessage
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil
	}
	return data
}

func diskCacheMTime() time.Time {
	path := modelsDevCachePath()
	if path == "" {
		return time.Time{}
	}
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func saveModelsDevDiskCache(data map[string]json.RawMessage) {
	path := modelsDevCachePath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	// Write atomically via temp file + rename.
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return
	}
	if err := json.NewEncoder(f).Encode(data); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return
	}
	f.Close()
	os.Rename(tmpPath, path)
}

// --- Network fetch ---

func fetchModelsDevNetwork() (map[string]json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsDevURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch models.dev: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("models.dev returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var data map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parse models.dev response: %w", err)
	}

	return data, nil
}
