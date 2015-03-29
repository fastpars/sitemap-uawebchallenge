package pagePraser

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

type Result struct {
	LastModified time.Time
	URLs         []*url.URL
}

func Get(u *url.URL) (Result, error) {
	result := Result{
		URLs: []*url.URL{},
	}

	r, err := http.Get(u.String())
	if err != nil {
		return result, err
	}
	defer r.Body.Close()

	doc, err := html.Parse(r.Body)
	if err != nil {
		return result, err
	}
	findLinks(doc, &result)

	lastModified := r.Header.Get("Last-Modified")
	if lastModified != "" {
		t, err := time.Parse(time.RFC1123, lastModified)
		if err == nil {
			result.LastModified = t
		}
	}
	return result, nil
}

func findLinks(n *html.Node, r *Result) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				u, err := url.Parse(a.Val)
				if err == nil {
					r.URLs = append(r.URLs, u)
				}
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findLinks(c, r)
	}
}

func GetLastmod(u string) (t time.Time) {
	r, err := http.Get(u)
	if err != nil {
		return
	}
	r.Body.Close()

	lastModified := r.Header.Get("Last-Modified")
	if lastModified != "" {
		t, err := time.Parse(time.RFC1123, lastModified)
		if err == nil {
			return t
		}
	}
	return
}
