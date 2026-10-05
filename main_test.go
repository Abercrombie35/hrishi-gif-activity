package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, target string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	handleIndex(rec, httptest.NewRequest(http.MethodGet, target, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d", target, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("GET %s: content type %q", target, ct)
	}
	return rec.Body.String()
}

func TestCount(t *testing.T) {
	cases := map[string]int{
		"/":                     1,
		"/?count=1":             1,
		"/?count=2":             2,
		"/?count=3":             3,
		"/?count=0":             1,
		"/?count=4":             1,
		"/?count=-2":            1,
		"/?count=banana":        1,
		"/?count=3&animals=all": 3,
	}
	for target, want := range cases {
		if got := strings.Count(get(t, target), `class="gif"`); got != want {
			t.Errorf("GET %s: %d GIFs, want %d", target, got, want)
		}
	}
}

func TestCatsByDefault(t *testing.T) {
	if len(catGIFs) == 0 || len(catGIFs) >= len(allGIFs) {
		t.Fatalf("cat filter kept %d of %d GIFs", len(catGIFs), len(allGIFs))
	}
	for i := 0; i < 50; i++ {
		body := get(t, "/?count=3")
		for _, line := range strings.Split(body, "\n") {
			if strings.Contains(line, `class="gif"`) {
				lower := strings.ToLower(line)
				if !strings.Contains(lower, "cat") && !strings.Contains(lower, "kitt") {
					t.Fatalf("non-cat GIF in default view: %s", line)
				}
			}
		}
	}
}

func TestPageContent(t *testing.T) {
	body := get(t, "/")
	for _, want := range []string{"<h1>", "<h2>", "<h3>", logoURL, `name="count"`, `method="get"`} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing %q", want)
		}
	}
}

func TestPickRandomNoDuplicates(t *testing.T) {
	for i := 0; i < 100; i++ {
		seen := map[string]bool{}
		for _, u := range pickRandom(allGIFs, 3) {
			if seen[u] {
				t.Fatalf("duplicate GIF %s", u)
			}
			seen[u] = true
		}
	}
}

func TestUnknownPath(t *testing.T) {
	rec := httptest.NewRecorder()
	handleIndex(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rec.Code)
	}
}
