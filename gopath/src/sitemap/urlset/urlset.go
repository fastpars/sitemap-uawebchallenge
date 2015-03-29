package urlset

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/url"

	"sitemap/pagePraser"
)

const MAX_SCAN_QUEUE = 100
const MAX_LAST_MODIFY_QUEUE = 100

const LastModifyLayout = "2006-01-02"

var (
	ErrAlreadyAdded    = errors.New("URL already added")
	ErrInvalidScheme   = errors.New("Invalid scheme")
	ErrInvalidURL      = errors.New("Invalid URL")
	ErrMissingURLHost  = errors.New("Missing URL Host")
	ErrNotQualifiedURL = errors.New("Not qualified URL")
)

type Urlset struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    URLs     `xml:"url"`

	rootURL              *url.URL
	maxLevel             uint
	scanQueue            chan *url.URL
	lastModifyQueue      chan *URL
	countScanQueue       uint
	countLastModifyQueue uint
	doneCh               chan struct{}
	done                 bool
}

type URLs map[string]*URL

type URL struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`

	level uint
}

func (urls URLs) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	for _, u := range urls {
		e.EncodeElement(u, start)
	}
	return nil
}

func validUrlset(u *url.URL) bool {
	return (u.Scheme == "http" || u.Scheme == "https") && u.Opaque == ""
}

func validUrl(u *url.URL) bool {
	return u.Opaque == ""
}

func MakeUrlset(u *url.URL, maxLevel uint) (*Urlset, error) {
	us := &Urlset{
		URLs:            URLs{},
		maxLevel:        maxLevel,
		scanQueue:       make(chan *url.URL, MAX_SCAN_QUEUE),
		lastModifyQueue: make(chan *URL, MAX_LAST_MODIFY_QUEUE),
		doneCh:          make(chan struct{}),
	}
	if !validUrlset(u) {
		return us, ErrInvalidURL
	}
	if u.Host == "" {
		return us, ErrMissingURLHost
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	us.rootURL = u
	us.scanQueue <- u
	us.queueAdd()

	us.URLs[u.String()] = &URL{Loc: u.String(), level: 0}

	return us, nil
}

func (us *Urlset) add(u *url.URL, level uint) error {
	if !validUrl(u) {
		return ErrInvalidURL
	}
	if u.Scheme == "" {
		u.Scheme = us.rootURL.Scheme
	}
	if u.Host == "" {
		u.Host = us.rootURL.Host
	}
	if u.Path == "/" {
		u.Path = ""
	}
	u.Fragment = ""
	if u.Host != us.rootURL.Host || u.Scheme != us.rootURL.Scheme {
		return ErrNotQualifiedURL
	}
	rawURL := u.String()
	if _, ok := us.URLs[rawURL]; ok {
		return ErrAlreadyAdded
	}
	us.URLs[rawURL] = &URL{Loc: u.String(), level: level}
	us.scanQueue <- u
	us.queueAdd()

	return nil
}

func (us *Urlset) Scan() {
	var u *url.URL

	for {
		select {
		case u = <-us.scanQueue:
		case <-us.doneCh:
			us.scanLastLevel()
			us.done = true
			return
		}
		go us.scan(u)
	}
}

func (us *Urlset) scanLastLevel() {
	go func() {
		for _, u := range us.URLs {
			if u.level == us.maxLevel {
				us.lastModifyQueue <- u
			}
		}
		close(us.lastModifyQueue)
	}()

	lastModifiedDone := make(chan struct{})
	for u := range us.lastModifyQueue {
		us.countLastModifyQueue++
		go u.LoadLastModified(lastModifiedDone)
	}
	for ; us.countLastModifyQueue > 0; us.countLastModifyQueue-- {
		<-lastModifiedDone
	}
}

func (u *URL) LoadLastModified(done chan struct{}) {
	lastModified := pagePraser.GetLastmod(u.Loc)
	if !lastModified.IsZero() {
		u.Lastmod = lastModified.Format(LastModifyLayout)
	}
	done <- struct{}{}
}

func (us *Urlset) queueAdd() {
	us.countScanQueue++
}

func (us *Urlset) queueDone() {
	us.countScanQueue--
	if us.countScanQueue == 0 {
		us.doneCh <- struct{}{}
	}
}

func (us *Urlset) scan(u *url.URL) error {
	defer us.queueDone()
	rawURL := u.String()
	r, err := pagePraser.Get(u)

	if err != nil {
		return err
	}
	if !r.LastModified.IsZero() {
		us.URLs[rawURL].Lastmod = r.LastModified.Format(LastModifyLayout)
	}

	nextLevel := us.URLs[rawURL].level + 1
	if nextLevel > us.maxLevel {
		return nil
	}

	for _, u := range r.URLs {
		us.add(u, nextLevel)
	}
	return nil
}

type UrlsetStat struct {
	RootURL         string `json:"rootURL"`
	MaxLevel        uint   `json:"maxLevel"`
	FoundURLs       int    `json:"foundURLs"`
	ScanQueue       uint   `json:"scanQueue"`
	LastModifyQueue uint   `json:"lastModifyQueue"`
	Done            bool   `json:"done"`
}

func (us *Urlset) Statistic(w io.Writer) {
	stat := UrlsetStat{
		RootURL:         us.rootURL.String(),
		MaxLevel:        us.maxLevel,
		FoundURLs:       len(us.URLs),
		ScanQueue:       us.countScanQueue,
		LastModifyQueue: us.countLastModifyQueue,
		Done:            us.done,
	}
	e := json.NewEncoder(w)
	e.Encode(stat)
}

func (us *Urlset) Xml(w io.Writer) {
	e := xml.NewEncoder(w)
	w.Write([]byte(xml.Header))
	e.Encode(us)
}
