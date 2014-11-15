package cmd

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"code.google.com/p/go.net/html"
	"github.com/PuerkitoBio/goquery"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "Start Flatten web service",
	Description: "",
	Action:      runWeb,
	Flags:       []cli.Flag{},
}

var tmpls map[string]*template.Template

const historyLength = 10

var history []string

func buildTemplates() {
	tmpls = make(map[string]*template.Template)
	var err error
	index_tmpl, err := template.ParseFiles("templates/base.tmpl", "templates/index.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmpls["index"] = index_tmpl

	about_tmpl, err := template.ParseFiles("templates/base.tmpl", "templates/about.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmpls["about"] = about_tmpl

	content_tmpl, err := template.ParseFiles("templates/base.tmpl", "templates/content.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmpls["content"] = content_tmpl
}

type Content struct {
	Title   string
	Entries []*Entry
	Error   string
}

type Entry struct {
	Author string
	Score  string
	Body   template.HTML
}

func scrapeHabrahabr(url string) *Content {

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	entries := make([]*Entry, 0)

	title := doc.Find("h1.title").First().Text()

	doc.Find("#comments .comment_item").Each(func(i int, s *goquery.Selection) {
		author := s.Find(".username").First().Text()
		score := s.Find(".mark .score").First().Text()
		content, err := s.Find(".message").First().Html()
		if err != nil {
			return
		}
		entry := &Entry{
			Author: author,
			Score:  score,
			Body:   template.HTML(html.UnescapeString(content)),
		}
		entries = append(entries, entry)
	})

	content := &Content{
		Title:   title,
		Entries: entries,
	}

	return content

}

func scrapeHackerNews(url string) *Content {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	entries := make([]*Entry, 0)

	title := doc.Find("td.title").First().Text()

	doc.Find("td.default").Each(func(i int, s *goquery.Selection) {
		author := s.Find(".comhead").Find("a").First().Text()
		content, err := s.Find(".comment").Find("font").First().Html()
		if err != nil {
			return
		}
		entry := &Entry{
			Author: author,
			Body:   template.HTML(html.UnescapeString(content)),
		}
		entries = append(entries, entry)
	})

	content := &Content{
		Title:   title,
		Entries: entries,
	}

	return content
}

func scrapeReddit(url string) *Content {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	entries := make([]*Entry, 0)

	title := doc.Find(".thing").First().Find(".title").First().Text()

	doc.Find(".commentarea .thing").Each(func(i int, s *goquery.Selection) {
		author := s.Find(".author").First().Text()
		score := s.Find(".score.unvoted").First().Text()
		content, err := s.Find(".usertext-body").Html()
		if err != nil {
			return
		}
		entry := &Entry{
			Author: author,
			Score:  score,
			Body:   template.HTML(html.UnescapeString(content)),
		}
		entries = append(entries, entry)
	})

	content := &Content{
		Title:   title,
		Entries: entries,
	}

	return content
}

func scrape(urlToProxy string) *Content {
	parsedUrl, err := url.Parse(urlToProxy)
	if err != nil {
		log.Fatal(err)
	}
	var content *Content
	if strings.HasSuffix(parsedUrl.Host, "reddit.com") {
		content = scrapeReddit(urlToProxy)
	} else if strings.HasSuffix(parsedUrl.Host, "news.ycombinator.com") {
		content = scrapeHackerNews(urlToProxy)
	} else if strings.HasSuffix(parsedUrl.Host, "habrahabr.ru") {
		content = scrapeHabrahabr(urlToProxy)
	} else {
		content = &Content{
			Error: "URL not supported yet, pull request can fix this",
		}
	}
	return content
}

func runWeb(*cli.Context) {

	history = make([]string, historyLength)

	buildTemplates()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		tmpls["index"].Execute(w, history)
	})

	mux.HandleFunc("/about/", func(w http.ResponseWriter, req *http.Request) {
		tmpls["about"].Execute(w, "")
	})

	mux.HandleFunc("/content/", func(w http.ResponseWriter, req *http.Request) {
		urlToProxy := req.FormValue("url")
		var content *Content
		if len(urlToProxy) == 0 {
			content = &Content{
				Error: "URL required",
			}
		} else {
			content = scrape(urlToProxy)
		}
		if content.Error == "" {
			history = append([]string{urlToProxy}, history...)
			history = history[:historyLength]
		}
		tmpls["content"].Execute(w, content)
	})

	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":3000")
}
