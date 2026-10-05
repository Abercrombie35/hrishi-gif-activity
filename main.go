package main

import (
	_ "embed"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

// Custom logo, hosted in a public repo.
const logoURL = "https://raw.githubusercontent.com/Abercrombie35/spin-the-blade/main/public/images/som_trivia_cat_logo.png"

const maxCount = 3

//go:embed gifs.txt
var gifList string

var (
	allGIFs = parseGIFs(gifList)
	catGIFs = filterCats(allGIFs)
)

var page = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SOM Trivia Cats</title>
  <style>
    body { margin: 0; padding: 2rem 1rem; background: #fdf0d5; color: #003049;
           font-family: system-ui, sans-serif; text-align: center; }
    h1 { margin-bottom: 0.25rem; }
    h2 { margin-top: 0; font-weight: 500; }
    h3 { color: #c1121f; }
    .logo { width: 220px; max-width: 60vw; }
    form { margin: 1.5rem auto; display: flex; gap: 0.75rem; justify-content: center;
           align-items: center; flex-wrap: wrap; }
    select, button { font-size: 1rem; padding: 0.35rem 0.6rem; }
    .gifs { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; }
    .gif { max-width: 320px; width: 100%; border-radius: 12px;
           box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2); }
  </style>
</head>
<body>
  <h1>Welcome to SOM Trivia</h1>
  <h2>Where every right answer earns you a cat</h2>
  <h3>Meet our mascot, then claim your reward</h3>
  <img class="logo" src="{{.Logo}}" alt="SOM Trivia cat logo">

  <form method="get" action="/">
    <label>How many GIFs?
      <select name="count">
        {{range .Counts}}<option value="{{.}}"{{if eq . $.Count}} selected{{end}}>{{.}}</option>{{end}}
      </select>
    </label>
    <label>Show
      <select name="animals">
        <option value="cats"{{if not .AllAnimals}} selected{{end}}>Cats only</option>
        <option value="all"{{if .AllAnimals}} selected{{end}}>All animals</option>
      </select>
    </label>
    <button type="submit">Show me</button>
  </form>

  <div class="gifs">
    {{range .GIFs}}<img class="gif" src="{{.}}" alt="A cute animal">
    {{end}}
  </div>
</body>
</html>
`))

type pageData struct {
	Logo       string
	Counts     []int
	Count      int
	AllAnimals bool
	GIFs       []string
}

func main() {
	http.HandleFunc("/", handleIndex)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Open http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	query := r.URL.Query()
	count := parseCount(query.Get("count"))
	allAnimals := query.Get("animals") == "all"

	pool := catGIFs
	if allAnimals {
		pool = allGIFs
	}

	data := pageData{
		Logo:       logoURL,
		Counts:     []int{1, 2, 3},
		Count:      count,
		AllAnimals: allAnimals,
		GIFs:       pickRandom(pool, count),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Execute(w, data); err != nil {
		log.Printf("render page: %v", err)
	}
}

// parseCount returns the requested number of GIFs, or 1 when the value is
// missing or invalid.
func parseCount(raw string) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxCount {
		return 1
	}
	return n
}

// pickRandom returns n different random entries from urls.
func pickRandom(urls []string, n int) []string {
	if n > len(urls) {
		n = len(urls)
	}
	picked := make([]string, 0, n)
	for _, i := range rand.Perm(len(urls))[:n] {
		picked = append(picked, urls[i])
	}
	return picked
}

func parseGIFs(list string) []string {
	var urls []string
	for _, line := range strings.Split(list, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			urls = append(urls, line)
		}
	}
	return urls
}

// filterCats keeps GIFs whose file names mention a cat or kitten.
func filterCats(urls []string) []string {
	var cats []string
	for _, u := range urls {
		name := strings.ToLower(path.Base(u))
		if strings.Contains(name, "cat") || strings.Contains(name, "kitt") {
			cats = append(cats, u)
		}
	}
	return cats
}
