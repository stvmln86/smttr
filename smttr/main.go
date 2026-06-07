package main

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Post struct {
	Body string
	Time time.Time
	Tone string
}

var (
	Conf  map[string]string
	Posts []*Post
)

func try(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// parse conf
	file, err := os.Open("confs/conf.json")
	try(err)
	try(json.NewDecoder(file).Decode(&Conf))
	log.Printf("parsed conf for %s", Conf["name"])

	// discover posts
	globs, err := filepath.Glob("posts/????/??/*.md")
	try(err)
	log.Printf("discovered %d posts", len(globs))

	// parse posts
	for _, orig := range globs {
		// read post
		bytes, err := os.ReadFile(orig)
		try(err)

		// separate post
		var lines []string
		var pairs = make(map[string]string)
		for line := range strings.SplitSeq(string(bytes), "\n") {
			if strings.HasPrefix(line, ".") {
				if attr, valu, okay := strings.Cut(line, " "); okay {
					pairs[attr] = strings.TrimSpace(valu)
				}
			} else {
				lines = append(lines, line)
			}
		}

		// parse post values
		body := strings.Join(lines, "\n")
		tone := strings.ToLower(pairs[".tone"])
		tobj, err := time.Parse("2006-01-02 15:04:05", pairs[".time"])
		try(err)

		// append post
		Posts = append(Posts, &Post{body, tobj, tone})
		log.Printf("parsed post %q", tobj)
	}

	// sort posts
	slices.SortFunc(Posts, func(a, b *Post) int { return a.Time.Compare(b.Time) })
	log.Printf("sorted %d posts", len(Posts))

	// prepare target directory
	os.Mkdir("targs", 0740) // ignored error

	// render index page
	tpls := template.Must(template.ParseFiles(
		"files/html/_base.html",
		"files/html/index.html",
	))

	data := struct {
		Conf  map[string]string
		Posts []*Post
	}{Conf, Posts}

	file, err = os.OpenFile("targs/index.html", os.O_CREATE|os.O_WRONLY, 0640)
	try(err)
	try(tpls.ExecuteTemplate(file, "_base.html", data))
	try(file.Close())
	log.Printf("rendered index template")
}
