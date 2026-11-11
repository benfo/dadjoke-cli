package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/benfourie/daily-dadjoke/internal/joke"
)

func TestRootCmd_OrthogonalFlagsAndMaxViews(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")
	t.Setenv("DADJOKE_CACHE_PATH", cachePath)

	today := joke.GetTodayKey(time.Now())
	maxViews := 2
	joke.WriteCache(joke.CachedJokeRecord{
		Date:     today,
		Joke:     "I'm reading a book on anti-gravity. It's impossible to put down.",
		Views:    2,
		MaxViews: &maxViews,
	}, cachePath)

	// 3rd view with --max-views 2 (like shell startup hook) -> hidden
	root1 := NewRootCmd()
	buf1 := new(bytes.Buffer)
	root1.SetOut(buf1)
	root1.SetArgs([]string{"--max-views", "2"})
	if err := root1.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf1.String() != "" {
		t.Errorf("expected empty output when --max-views 2 is exceeded, got %q", buf1.String())
	}

	// 4th view with --plain --max-views 2 -> also hidden (--plain is strictly formatting!)
	root2 := NewRootCmd()
	buf2 := new(bytes.Buffer)
	root2.SetOut(buf2)
	root2.SetArgs([]string{"--plain", "--max-views", "2"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf2.String() != "" {
		t.Errorf("expected --plain --max-views 2 to respect view limit, got %q", buf2.String())
	}

	// Manual run (`dadjoke` with no --max-views) -> always prints boxed joke!
	root3 := NewRootCmd()
	buf3 := new(bytes.Buffer)
	root3.SetOut(buf3)
	root3.SetArgs([]string{})
	if err := root3.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf3.String(), "impossible to put down") || !strings.Contains(buf3.String(), "╭") {
		t.Errorf("expected manual run to print boxed joke, got:\n%s", buf3.String())
	}

	// Manual run (`dadjoke --plain` with no --max-views) -> always prints plain joke!
	root4 := NewRootCmd()
	buf4 := new(bytes.Buffer)
	root4.SetOut(buf4)
	root4.SetArgs([]string{"--plain"})
	if err := root4.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf4.String()) != "I'm reading a book on anti-gravity. It's impossible to put down." {
		t.Errorf("expected manual --plain run to print plain joke, got %q", buf4.String())
	}
}

func TestRootCmd_RejectsNegativeMaxViews(t *testing.T) {
	root := NewRootCmd()
	root.SetArgs([]string{"--max-views", "-1"})
	if err := root.Execute(); err == nil {
		t.Fatalf("expected error for negative --max-views, got nil")
	}
}
