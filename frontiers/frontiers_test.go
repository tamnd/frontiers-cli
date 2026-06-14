package frontiers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/frontiers-cli/frontiers"
)

func papersPayload(items []map[string]any) []byte {
	b, _ := json.Marshal(map[string]any{
		"message": map[string]any{
			"total-results": len(items),
			"items":         items,
		},
	})
	return b
}

func TestRecent(t *testing.T) {
	payload := papersPayload([]map[string]any{
		{
			"DOI":                    "10.3389/fnins.2026.000001",
			"title":                  []string{"Neural Circuits in Cognitive Processing"},
			"author":                 []map[string]any{{"given": "Alice", "family": "Smith"}},
			"published":              map[string]any{"date-parts": [][]int{{2026, 6}}},
			"type":                   "journal-article",
			"container-title":        []string{"Frontiers in Neuroscience"},
			"is-referenced-by-count": 3,
		},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	cfg := frontiers.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := frontiers.NewClient(cfg)
	papers, err := c.Recent(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 1 {
		t.Fatalf("got %d papers, want 1", len(papers))
	}
	if papers[0].Title != "Neural Circuits in Cognitive Processing" {
		t.Errorf("title = %q", papers[0].Title)
	}
	if papers[0].Rank != 1 {
		t.Errorf("rank = %d, want 1", papers[0].Rank)
	}
	if papers[0].Published != "2026-06" {
		t.Errorf("published = %q, want 2026-06", papers[0].Published)
	}
	if papers[0].URL != "https://doi.org/10.3389/fnins.2026.000001" {
		t.Errorf("url = %q", papers[0].URL)
	}
}

func TestRecentLimit(t *testing.T) {
	items := make([]map[string]any, 5)
	for i := range items {
		items[i] = map[string]any{
			"DOI":                    "10.3389/test",
			"title":                  []string{"Test Paper"},
			"published":              map[string]any{"date-parts": [][]int{{2025}}},
			"is-referenced-by-count": 0,
		}
	}
	payload := papersPayload(items)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	cfg := frontiers.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := frontiers.NewClient(cfg)
	papers, err := c.Recent(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 3 {
		t.Fatalf("got %d papers, want 3 (limit applied)", len(papers))
	}
}

func TestSearch(t *testing.T) {
	payload := papersPayload([]map[string]any{
		{
			"DOI":                    "10.3389/fpsyg.2024.001",
			"title":                  []string{"Deep Learning for Image Classification"},
			"author":                 []map[string]any{{"given": "Bob", "family": "Jones"}},
			"published":              map[string]any{"date-parts": [][]int{{2024, 3}}},
			"container-title":        []string{"Frontiers in Psychology"},
			"is-referenced-by-count": 10,
		},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	cfg := frontiers.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := frontiers.NewClient(cfg)
	papers, err := c.Search(context.Background(), "deep learning", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 1 {
		t.Fatalf("got %d papers, want 1", len(papers))
	}
	if papers[0].Authors != "Jones B" {
		t.Errorf("authors = %q, want 'Jones B'", papers[0].Authors)
	}
}
