// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is a blog server for articles written in present format.
// It powers jtwsm.net.
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"jtwsm.net/code/pkg/atom"
	"jtwsm.net/code/pkg/present"
	"jtwsm.net/code/pkg/tmplfunc"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	hostname     = "blog.jtwsm.net"
	baseURL      = "http://" + hostname
	homeArticles = 5  // articles to display on the home page
	feedArticles = 10 // articles to include in Atom and JSON feeds
)

var validJSONPFunc = regexp.MustCompile(`(?i)^[a-z_][a-z0-9_.]*$`)

// Doc represents an article, adorned with presentation data:
// its absolute path and related articles.
type Doc struct {
	*present.Doc
	Permalink    string
	Path         string
	Related      []*Doc
	Newer, Older *Doc
	HTML         template.HTML // rendered article
}

// Server implements an http.Handler that serves blog articles.
type Server struct {
	docs     []*Doc
	tags     []string
	docPaths map[string]*Doc
	docTags  map[string][]*Doc
	template struct {
		home, index, article, doc *template.Template
	}
	atomFeed []byte // pre-rendered Atom feed
	jsonFeed []byte // pre-rendered JSON feed
	content  http.Handler
}

// NewServer constructs a new Server, serving articles from the specified
// contentPath generated from templates from templatePath.
func NewServer(contentPath, templatePath string) (*Server, error) {
	present.PlayEnabled = true

	root := filepath.Join(templatePath, "root.tmpl")
	parse := func(name string) (*template.Template, error) {
		t := template.New("").Funcs(funcMap)
		return t.ParseFiles(root, filepath.Join(templatePath, name))
	}

	s := &Server{}

	// Parse templates.
	var err error
	s.template.home, err = parse("home.tmpl")
	if err != nil {
		return nil, err
	}
	s.template.index, err = parse("index.tmpl")
	if err != nil {
		return nil, err
	}
	s.template.article, err = parse("article.tmpl")
	if err != nil {
		return nil, err
	}
	p := present.Template().Funcs(funcMap)
	s.template.doc, err = p.ParseFiles(filepath.Join(templatePath, "doc.tmpl"))
	if err != nil {
		return nil, err
	}

	// Load content.
	err = s.loadDocs(filepath.Clean(contentPath))
	if err != nil {
		return nil, err
	}

	err = s.renderAtomFeed()
	if err != nil {
		return nil, err
	}

	err = s.renderJSONFeed()
	if err != nil {
		return nil, err
	}

	// Set up content file server.
	s.content = http.FileServer(http.Dir(contentPath))

	return s, nil
}

var funcMap = template.FuncMap{}

func init() {
	funcMap["equal"] = tmplfunc.FuncMap["equal"]
	funcMap["eq"] = tmplfunc.FuncMap["eq"]
	funcMap["lt"] = tmplfunc.FuncMap["lt"]
	funcMap["le"] = tmplfunc.FuncMap["le"]
	funcMap["gt"] = tmplfunc.FuncMap["gt"]
	funcMap["ge"] = tmplfunc.FuncMap["ge"]
	funcMap["add"] = tmplfunc.FuncMap["add"]
	funcMap["sub"] = tmplfunc.FuncMap["sub"]
}

// authors returns a comma-separated list of author names.
func authors(authors []present.Author) string {
	var b bytes.Buffer
	last := len(authors) - 1
	for i, a := range authors {
		if i > 0 {
			if i == last {
				b.WriteString(" and ")
			} else {
				b.WriteString(", ")
			}
		}
		b.WriteString(a.Name)
	}
	return b.String()
}

// loadDocs reads all content from the provided file system root, renders all
// the articles it finds, adds them to the Server's docs field, computes the
// denormalized docPaths, docTags, and tags fields, and populates the various
// helper fields (Next, Previous, Related) for each Doc.
func (s *Server) loadDocs(root string) error {
	// Read content into docs field.
	const ext = ".article"
	fn := func(p string, info os.FileInfo, err error) error {
		if filepath.Ext(p) != ext {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		d, err := present.Parse(f, p, 0)
		if err != nil {
			return err
		}
		html := new(bytes.Buffer)
		err = d.Render(html, s.template.doc)
		if err != nil {
			return err
		}
		p = "/" + p[len(root)+1:len(p)-len(ext)] // trim root and extension
		s.docs = append(s.docs, &Doc{
			Doc:       d,
			Path:      p,
			Permalink: baseURL + p,
			HTML:      template.HTML(html.String()),
		})
		return nil
	}
	err := filepath.Walk(root, fn)
	if err != nil {
		return err
	}
	sort.Sort(docsByTime(s.docs))

	// Pull out doc paths and tags and put in reverse-associating maps.
	s.docPaths = make(map[string]*Doc)
	s.docTags = make(map[string][]*Doc)
	for _, d := range s.docs {
		s.docPaths[d.Path] = d
		for _, t := range d.Tags {
			s.docTags[t] = append(s.docTags[t], d)
		}
	}

	// Pull out unique sorted list of tags.
	for t := range s.docTags {
		s.tags = append(s.tags, t)
	}
	sort.Strings(s.tags)

	// Set up presentation-related fields, Newer, Older, and Related.
	for _, doc := range s.docs {
		// Newer, Older: docs adjacent to doc
		for i := range s.docs {
			if s.docs[i] != doc {
				continue
			}
			if i > 0 {
				doc.Newer = s.docs[i-1]
			}
			if i+1 < len(s.docs) {
				doc.Older = s.docs[i+1]
			}
			break
		}

		// Related: all docs that share tags with doc.
		related := make(map[*Doc]bool)
		for _, t := range doc.Tags {
			for _, d := range s.docTags[t] {
				if d != doc {
					related[d] = true
				}
			}
		}
		for d := range related {
			doc.Related = append(doc.Related, d)
		}
		sort.Sort(docsByTime(doc.Related))
	}

	return nil
}

// renderAtomFeed generates an XML Atom feed and stores it in the Server's
// atomFeed field.
func (s *Server) renderAtomFeed() error {
	var updated time.Time
	if len(s.docs) > 1 {
		updated = s.docs[0].Time
	}
	feed := atom.Feed{
		Title:   "某某人的博客",
		ID:      "tag:" + hostname + ",2013:" + hostname,
		Updated: atom.Time(updated),
		Link: []atom.Link{{
			Rel:  "self",
			Href: baseURL + "/feed.atom",
		}},
	}
	for i, doc := range s.docs {
		if i >= feedArticles {
			break
		}
		e := &atom.Entry{
			Title: doc.Title,
			ID:    feed.ID + doc.Path,
			Link: []atom.Link{{
				Rel:  "alternate",
				Href: baseURL + doc.Path,
			}},
			Published: atom.Time(doc.Time),
			Updated:   atom.Time(doc.Time),
			Summary: &atom.Text{
				Type: "html",
				Body: summary(doc),
			},
			Content: &atom.Text{
				Type: "html",
				Body: string(doc.HTML),
			},
			Author: &atom.Person{
				Name: authors(doc.Authors),
			},
		}
		feed.Entry = append(feed.Entry, e)
	}
	data, err := xml.Marshal(&feed)
	if err != nil {
		return err
	}
	s.atomFeed = data
	return nil
}

type jsonItem struct {
	Title   string
	Link    string
	Time    time.Time
	Summary string
	Content string
	Author  string
}

// renderJSONFeed generates a JSON feed and stores it in the Server's jsonFeed
// field.
func (s *Server) renderJSONFeed() error {
	var feed []jsonItem
	for i, doc := range s.docs {
		if i >= feedArticles {
			break
		}
		item := jsonItem{
			Title:   doc.Title,
			Link:    baseURL + doc.Path,
			Time:    doc.Time,
			Summary: summary(doc),
			Content: string(doc.HTML),
			Author:  authors(doc.Authors),
		}
		feed = append(feed, item)
	}
	data, err := json.Marshal(feed)
	if err != nil {
		return err
	}
	s.jsonFeed = data
	return nil
}

// summary returns the first paragraph of text from the provided Doc.
func summary(d *Doc) string {
	if len(d.Sections) == 0 {
		return ""
	}
	sec := d.Sections[0].Inner.(*present.Section)
	for _, ep := range sec.Elems {
		e := ep.Inner.(present.Elem)
		text, ok := e.(*present.Text)
		if !ok || text.Pre {
			// skip everything but non-text elements
			continue
		}
		var buf bytes.Buffer
		for _, s := range text.Lines {
			buf.WriteString(string(present.Style(s)))
			buf.WriteByte('\n')
		}
		return buf.String()
	}
	return ""
}

// rootData encapsulates data destined for the root template.
type rootData struct {
	Doc  *Doc
	Data interface{}
}

// ServeHTTP servers either an article list or a single article.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		d rootData
		t *template.Template
	)
	switch p := r.URL.Path; p {
	case "/":
		d.Data = s.docs
		if len(s.docs) > homeArticles {
			d.Data = s.docs[:homeArticles]
		}
		t = s.template.home
	case "/index":
		d.Data = s.docs
		t = s.template.index
	case "/feed.atom", "/feeds/posts/default":
		w.Header().Set("Content-type", "application/atom+xml; charset=utf-8")
		w.Write(s.atomFeed)
		return
	case "/.json":
		if p := r.FormValue("jsonp"); validJSONPFunc.MatchString(p) {
			w.Header().Set("Content-type", "application/javascript; charset=utf-8")
			fmt.Fprintf(w, "%v(%s)", p, s.jsonFeed)
			return
		}
		w.Header().Set("Content-type", "application/json; charset=utf-8")
		w.Write(s.jsonFeed)
		return
	default:
		doc, ok := s.docPaths[p]
		if !ok {
			// Not a doc; try to just serve static content.
			s.content.ServeHTTP(w, r)
			return
		}
		d.Doc = doc
		t = s.template.article
	}
	err := t.ExecuteTemplate(w, "root", d)
	if err != nil {
		log.Println(err)
	}
}

// docsByTime implements sort.Interface, sorting Docs by their Time field.
type docsByTime []*Doc

func (s docsByTime) Len() int           { return len(s) }
func (s docsByTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s docsByTime) Less(i, j int) bool { return s[i].Time.After(s[j].Time) }
