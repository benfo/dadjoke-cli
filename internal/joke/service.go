package joke

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// JokeAPIURL is the icanhazdadjoke endpoint.
	JokeAPIURL = "https://icanhazdadjoke.com/"
	// DefaultMaxDailyViews is the default number of terminal launches per day that display the joke.
	DefaultMaxDailyViews = 3
	// DefaultTimeout is the HTTP client timeout for fetching a joke.
	DefaultTimeout = 3 * time.Second
)

// CachedJokeRecord represents the JSON structure stored in ~/.dadjoke-cache.json.
type CachedJokeRecord struct {
	Date     string  `json:"date"`
	Joke     string  `json:"joke"`
	ID       *string `json:"id,omitempty"`
	CachedAt string  `json:"cachedAt,omitempty"`
	Views    int     `json:"views,omitempty"`
	MaxViews *int    `json:"maxViews,omitempty"`
}

// Source indicates where the returned joke came from.
type Source string

const (
	SourceCache           Source = "cache"
	SourceNetwork         Source = "network"
	SourceStaleCache      Source = "stale-cache"
	SourceOfflineFallback Source = "offline-fallback"
)

// Result contains the resolved joke and view metadata.
type Result struct {
	Joke     string
	Source   Source
	Date     string
	Views    int
	MaxViews int
	Hidden   bool
}

// RemoteJoke represents the JSON payload returned by icanhazdadjoke.com.
type RemoteJoke struct {
	ID   string `json:"id"`
	Joke string `json:"joke"`
}

// HTTPDoer abstracts http.Client for testing.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// DailyJokeOptions configures GetDailyJoke behavior.
type DailyJokeOptions struct {
	ForceRefresh bool
	ForceShow    bool
	MaxViews     *int
	CachePath    string
	TodayKey     string
	Client       HTTPDoer
	APIURL       string
	Timeout      time.Duration
}

// FetchOptions configures remote joke fetching.
type FetchOptions struct {
	Client  HTTPDoer
	APIURL  string
	Timeout time.Duration
}

// GetTodayKey formats time t as YYYY-MM-DD in local time.
func GetTodayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

// GetDefaultCachePath returns DADJOKE_CACHE_PATH if set, or ~/.dadjoke-cache.json.
func GetDefaultCachePath() string {
	if envPath := strings.TrimSpace(os.Getenv("DADJOKE_CACHE_PATH")); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".dadjoke-cache.json"
	}
	return filepath.Join(home, ".dadjoke-cache.json")
}

// ReadCache reads and validates the cached joke record from disk.
func ReadCache(cachePath string) *CachedJokeRecord {
	if cachePath == "" {
		cachePath = GetDefaultCachePath()
	}
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}
	var record CachedJokeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil
	}
	if strings.TrimSpace(record.Date) == "" || strings.TrimSpace(record.Joke) == "" {
		return nil
	}
	return &record
}

// WriteCache writes the cached joke record to disk, creating parent directories if needed.
func WriteCache(record CachedJokeRecord, cachePath string) {
	if cachePath == "" {
		cachePath = GetDefaultCachePath()
	}
	dir := filepath.Dir(cachePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(cachePath, data, 0o644)
}

// HasCachedJokeForToday returns true if a valid joke is already cached for todayKey.
func HasCachedJokeForToday(cachePath, todayKey string) bool {
	if cachePath == "" {
		cachePath = GetDefaultCachePath()
	}
	if todayKey == "" {
		todayKey = GetTodayKey(time.Now())
	}
	cached := ReadCache(cachePath)
	return cached != nil && cached.Date == todayKey && cached.Joke != ""
}

// FetchRemoteDadJoke fetches a random dad joke from the API.
func FetchRemoteDadJoke(opts FetchOptions) *RemoteJoke {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	apiURL := opts.APIURL
	if apiURL == "" {
		apiURL = JokeAPIURL
	}
	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "DailyDadJokeCLI (https://github.com/benfourie/daily-dadjoke)")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}

	var payload RemoteJoke
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil
	}
	trimmed := strings.TrimSpace(payload.Joke)
	if trimmed == "" {
		return nil
	}
	return &RemoteJoke{
		ID:   payload.ID,
		Joke: trimmed,
	}
}

// GetDailyJoke retrieves today's joke from cache, network, stale cache, or offline fallback.
func GetDailyJoke(opts DailyJokeOptions) Result {
	cachePath := opts.CachePath
	if cachePath == "" {
		cachePath = GetDefaultCachePath()
	}
	todayKey := opts.TodayKey
	if todayKey == "" {
		todayKey = GetTodayKey(time.Now())
	}

	cached := ReadCache(cachePath)
	effectiveMaxViews := 0
	if opts.MaxViews != nil {
		effectiveMaxViews = *opts.MaxViews
	}

	if !opts.ForceRefresh && cached != nil && cached.Date == todayKey && cached.Joke != "" {
		prevViews := cached.Views
		if prevViews <= 0 {
			prevViews = 1
		}
		updatedViews := prevViews + 1
		updatedRecord := *cached
		updatedRecord.Views = updatedViews
		if opts.MaxViews != nil {
			maxViewsCopy := effectiveMaxViews
			updatedRecord.MaxViews = &maxViewsCopy
		}
		WriteCache(updatedRecord, cachePath)

		hidden := !opts.ForceShow && effectiveMaxViews > 0 && updatedViews > effectiveMaxViews
		return Result{
			Joke:     cached.Joke,
			Source:   SourceCache,
			Date:     cached.Date,
			Views:    updatedViews,
			MaxViews: effectiveMaxViews,
			Hidden:   hidden,
		}
	}

	remote := FetchRemoteDadJoke(FetchOptions{
		Client:  opts.Client,
		APIURL:  opts.APIURL,
		Timeout: opts.Timeout,
	})
	if remote != nil && remote.Joke != "" {
		var idPtr *string
		if remote.ID != "" {
			idCopy := remote.ID
			idPtr = &idCopy
		}
		maxViewsCopy := effectiveMaxViews
		record := CachedJokeRecord{
			Date:     todayKey,
			Joke:     remote.Joke,
			ID:       idPtr,
			CachedAt: time.Now().UTC().Format(time.RFC3339),
			Views:    1,
			MaxViews: &maxViewsCopy,
		}
		WriteCache(record, cachePath)
		return Result{
			Joke:     remote.Joke,
			Source:   SourceNetwork,
			Date:     todayKey,
			Views:    1,
			MaxViews: effectiveMaxViews,
			Hidden:   false,
		}
	}

	if cached != nil && cached.Joke != "" {
		return Result{
			Joke:     cached.Joke,
			Source:   SourceStaleCache,
			Date:     cached.Date,
			Views:    1,
			MaxViews: effectiveMaxViews,
			Hidden:   false,
		}
	}

	return Result{
		Joke:     GetFallbackJoke(todayKey),
		Source:   SourceOfflineFallback,
		Date:     todayKey,
		Views:    1,
		MaxViews: effectiveMaxViews,
		Hidden:   false,
	}
}

// GetRandomJoke fetches a fresh random dad joke without reading or updating the daily cache.
func GetRandomJoke(opts FetchOptions) Result {
	remote := FetchRemoteDadJoke(opts)
	if remote != nil && remote.Joke != "" {
		return Result{
			Joke:   remote.Joke,
			Source: SourceNetwork,
		}
	}
	seed := fmt.Sprintf("rand-%d", rand.Int63())
	return Result{
		Joke:   GetFallbackJoke(seed),
		Source: SourceOfflineFallback,
	}
}
