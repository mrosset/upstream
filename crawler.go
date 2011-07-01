package main

import (
	"bytes"
	"fmt"
	"http"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	timer "github.com/str1ngs/gotimer"
)

var (
	client = new(http.Client)
)

type Crawler struct {
	templates map[string]*Template
}

func NewCrawler(srcpath string) (*Crawler, os.Error) {
	ts, err := GetTemplates(srcpath)
	if err != nil {
		return nil, err
	}
	return &Crawler{ts}, nil
}

func (c *Crawler) Start() os.Error {
	log.Println("staring", "crawl")
	defer timer.From(timer.Now())
	var (
		done = 0
	)
	for _, t := range c.templates {
		if t.Distfiles == "" {
			continue
		}
		c.crawl(t)
		done++
	}
	return nil
}

func (c *Crawler) Crawl(key string) {
	c.crawl(c.templates[key])
}

func getParentUrl(t *Template) (string, string) {
	t.Distfiles = strings.Trim(t.Distfiles, " ")
	distfiles := strings.Split(t.Distfiles, " ")
	url, file := filepath.Split(distfiles[0])
	if url[:3] == "ftp" {
		url = "http" + url[3:]
	}
	return url, file
}

func (c *Crawler) crawl(t *Template) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(t.Pkgname, err)
		}
	}()
	rawurl, file := getParentUrl(t)
	url, err := http.ParseURL(rawurl)
	if err != nil {
		log.Println(t.Pkgname, err)
		return
	}
	// sourceforge sucks, make it suckless
	if url.Host == "downloads.sourceforge.net" {
		rawurl = fmt.Sprintf("http://www.sourceforge.net/projects/%s/files/", t.Pkgname)
	}
	res, err := client.Get(rawurl)
	if err != nil {
		log.Println(t.Pkgname, err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println(t.Pkgname, fmt.Errorf("HTTP status %s", res.Status), rawurl)
		return
	}
	rbuf := new(bytes.Buffer)
	_, err = io.Copy(rbuf, res.Body)
	if err != nil {
		log.Println(t.Pkgname, err)
		return
	}
	regex := `%s[a-z0-9\-\.]+(%s)`
	fext := filepath.Ext(file)
	reg := regexp.MustCompile(fmt.Sprintf(regex, t.Pkgname, fext[1:]))
	results := reg.FindAllString(string(rbuf.Bytes()), -1)
	if len(results) == 0 {
		log.Println(t.Pkgname, "no distfiles found on", rawurl)
		return
	}
	vregx := regxFromFile(file)
	latest := findLatest(results, vregx)
	if latest.Int > verInt(t.Version) {
		newlog, err := os.OpenFile("upstream_new.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Println(t.Pkgname, err)
		}
		defer newlog.Close()
		mw := io.MultiWriter(os.Stderr, newlog)
		fmt.Fprintf(mw, "%-20.20s upstream %10.10s vanilla %10.10s %s\n", t.Pkgname, latest.String, t.Version, rawurl)
	}
}

type Latest struct {
	Int    int
	String string
}

func findLatest(files []string, vregx string) *Latest {
	var (
		latest = &Latest{}
	)
	for _, file := range files {
		if vregx == regxFromFile(file) {
			sv := regexp.MustCompile(vregx).FindString(file)
			if verInt(sv) > latest.Int {
				latest.Int = verInt(sv)
				latest.String = sv
			}
		}
	}
	return latest
}

func verInt(ver string) int {
	ver = strings.Replace(ver, ".", "", -1)
	i, err := strconv.Atoi(ver)
	if err != nil {
		return 0
	}
	return i
}

func regxFromFile(file string) string {
	var (
		quad   = `[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`
		truple = `[0-9]+\.[0-9]+\.[0-9]+`
		double = `[0-9]+\.[0-9]+`
	)
	if ok, _ := regexp.MatchString(quad, file); ok {
		return quad
	}
	if ok, _ := regexp.MatchString(truple, file); ok {
		return truple
	}
	if ok, _ := regexp.MatchString(double, file); ok {
		return double
	}
	return ""
}
