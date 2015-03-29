package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"sitemap/urlset"
)

var urlsets = map[string]*urlset.Urlset{}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index, err := ioutil.ReadFile("www/index.html")
		if err != nil {
			return
		}

		tmpl, err := template.New("index").Parse(string(index))
		err = tmpl.Execute(w, "test")
		if err != nil {
			return
		}
	})

	http.HandleFunc("/generateSitemap", func(w http.ResponseWriter, r *http.Request) {
		rawHomeURL := r.FormValue("homeURL")

		levels, err := strconv.Atoi(r.FormValue("levels"))
		if err != nil || levels < 0 {
			return
		}

		homeURL, err := url.Parse(rawHomeURL)
		if err != nil {
			return
		}

		set, err := urlset.MakeUrlset(homeURL, uint(levels))
		if err != nil {
			return
		}

		hash := sha256.Sum256([]byte(time.Now().Format(time.RFC3339Nano) + "UA Web Challenge 2015"))
		token := hex.EncodeToString(hash[:])

		urlsets[token] = set
		go set.Scan()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":"%s"}`, token)
	})

	http.HandleFunc("/sitemapStatistic", func(w http.ResponseWriter, r *http.Request) {
		var set *urlset.Urlset
		var ok bool
		token := r.FormValue("token")
		if set, ok = urlsets[token]; !ok {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		set.Statistic(w)
	})

	http.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		var set *urlset.Urlset
		var ok bool
		token := r.URL.Query().Get("token")
		if set, ok = urlsets[token]; !ok {
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		set.Xml(w)
	})

	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./www/media"))))

	http.ListenAndServe(":8888", nil)
}
