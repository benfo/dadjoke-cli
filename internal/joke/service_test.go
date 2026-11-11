package joke

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func newJSONResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestGetDailyJoke_FetchesAndCachesOnCacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")
	today := "2026-11-11"

	calls := 0
	client := &mockHTTPClient{
		fn: func(req *http.Request) (*http.Response, error) {
			calls++
			if req.Header.Get("Accept") != "application/json" {
				t.Errorf("expected Accept: application/json, got %s", req.Header.Get("Accept"))
			}
			return newJSONResponse(200, `{"id":"abc123","joke":"Why did the Go gopher cross the road? To get to the other goroutine."}`), nil
		},
	}

	res := GetDailyJoke(DailyJokeOptions{
		CachePath: cachePath,
		TodayKey:  today,
		Client:    client,
	})

	if calls != 1 {
		t.Fatalf("expected 1 network call, got %d", calls)
	}
	if res.Source != SourceNetwork {
		t.Errorf("expected source %s, got %s", SourceNetwork, res.Source)
	}
	if res.Views != 1 || res.Hidden {
		t.Errorf("expected views=1 hidden=false, got views=%d hidden=%v", res.Views, res.Hidden)
	}

	// Subsequent call on the same day should hit cache and increment views
	res2 := GetDailyJoke(DailyJokeOptions{
		CachePath: cachePath,
		TodayKey:  today,
		Client:    client,
	})
	if calls != 1 {
		t.Fatalf("expected no additional network calls on cache hit, got %d", calls)
	}
	if res2.Source != SourceCache {
		t.Errorf("expected source %s, got %s", SourceCache, res2.Source)
	}
	if res2.Views != 2 || res2.Hidden {
		t.Errorf("expected views=2 hidden=false, got views=%d hidden=%v", res2.Views, res2.Hidden)
	}
}

func TestGetDailyJoke_HidesAfterMaxDailyViewsUnlessForced(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")
	today := "2026-11-11"

	client := &mockHTTPClient{
		fn: func(req *http.Request) (*http.Response, error) {
			return newJSONResponse(200, `{"id":"j1","joke":"Testing view limits!"}`), nil
		},
	}

	limit := 3
	// Views 1, 2, 3 should be visible when MaxViews = 3
	for i := 1; i <= 3; i++ {
		r := GetDailyJoke(DailyJokeOptions{CachePath: cachePath, TodayKey: today, Client: client, MaxViews: &limit})
		if r.Hidden {
			t.Fatalf("expected view %d to be visible, but was hidden", i)
		}
		if r.Views != i {
			t.Fatalf("expected view count %d, got %d", i, r.Views)
		}
	}

	// View 4 with MaxViews=3 should be hidden
	r4 := GetDailyJoke(DailyJokeOptions{CachePath: cachePath, TodayKey: today, Client: client, MaxViews: &limit})
	if !r4.Hidden {
		t.Fatalf("expected view 4 to be hidden after reaching maxViews=3")
	}

	// View 5 without MaxViews (manual CLI invocation) should always show
	rManual := GetDailyJoke(DailyJokeOptions{CachePath: cachePath, TodayKey: today, Client: client})
	if rManual.Hidden {
		t.Fatalf("expected manual run (no MaxViews) to always display joke")
	}

	// ForceShow with MaxViews=3 should show even past maxViews
	r5 := GetDailyJoke(DailyJokeOptions{CachePath: cachePath, TodayKey: today, Client: client, MaxViews: &limit, ForceShow: true})
	if r5.Hidden {
		t.Fatalf("expected ForceShow to display joke even past maxViews")
	}

	// MaxViews = 0 (unlimited) should also show
	zero := 0
	r6 := GetDailyJoke(DailyJokeOptions{CachePath: cachePath, TodayKey: today, Client: client, MaxViews: &zero})
	if r6.Hidden {
		t.Fatalf("expected MaxViews=0 (unlimited) to display joke")
	}
}

func TestGetDailyJoke_FallbackToStaleCacheAndOffline(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	failingClient := &mockHTTPClient{
		fn: func(req *http.Request) (*http.Response, error) {
			return newJSONResponse(503, "unavailable"), nil
		},
	}

	// No cache + failing network -> offline fallback
	resOffline := GetDailyJoke(DailyJokeOptions{
		CachePath: cachePath,
		TodayKey:  "2026-11-11",
		Client:    failingClient,
	})
	if resOffline.Source != SourceOfflineFallback || resOffline.Joke == "" {
		t.Fatalf("expected offline fallback joke, got %+v", resOffline)
	}

	// Seed yesterday's cache -> stale-cache fallback on network error
	WriteCache(CachedJokeRecord{
		Date:  "2026-11-10",
		Joke:  "Yesterday's cached dad joke.",
		Views: 2,
	}, cachePath)

	resStale := GetDailyJoke(DailyJokeOptions{
		CachePath: cachePath,
		TodayKey:  "2026-11-11",
		Client:    failingClient,
	})
	if resStale.Source != SourceStaleCache || resStale.Joke != "Yesterday's cached dad joke." {
		t.Fatalf("expected stale cache fallback, got %+v", resStale)
	}
}

func TestFormatJoke_PlainAndBoxed(t *testing.T) {
	plain := FormatJoke("What do you call a fake noodle? An impasta.", FormatOptions{Plain: true})
	if plain != "What do you call a fake noodle? An impasta." {
		t.Errorf("unexpected plain output: %q", plain)
	}

	boxed := FormatJoke("What do you call a fake noodle? An impasta.", FormatOptions{
		Plain: false,
		Title: "Today's Dad Joke",
	})
	if !strings.Contains(boxed, "Today's Dad Joke") || !strings.Contains(boxed, "An impasta.") {
		t.Errorf("expected boxed output to contain title and joke, got:\n%s", boxed)
	}
	if !strings.Contains(boxed, "╭") || !strings.Contains(boxed, "╯") {
		t.Errorf("expected rounded border characters in boxed output, got:\n%s", boxed)
	}
}
