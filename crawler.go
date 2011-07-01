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
	"runtime"
	"time"
)

var (
	client = new(http.Client)
	done   = 0
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
	tick := time.Tick(1e09 / 2)
	maxjobs := int32(100)
	for _, t := range c.templates {
		if t.Distfiles == "" {
			continue
		}
		jobs := runtime.Goroutines()
		if jobs >= maxjobs {
		wait:
			for {
				select {
				case <-tick:
					if runtime.Goroutines() < maxjobs {
						break wait
					}
				default:
					runtime.Gosched()
				}
			}
		}
		go c.crawl(t)
		done++
		fmt.Printf("\rjobs %3.3v done %v", runtime.Goroutines(), done)

	}
	for {
		if runtime.Goroutines() == 0 {
			break
		}
		fmt.Printf("\rjobs %3.3v done %v", runtime.Goroutines(), done)
		runtime.Gosched()
		<-tick
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
		rawurl = fmt.Sprintf("http://sourceforge.net/projects/%s/files/", t.Pkgname)
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
	for _, r := range results {
		log.Println(t.Pkgname, r)
	}
}

func fileToRegx(file string) string {
	rbuf := new(bytes.Buffer)
	for _, c := range file {
		_, err := strconv.Atoi(string(c))
		if err == nil {
			rbuf.WriteString("[0-9]")
			continue
		}
		rbuf.WriteString(string(c))
	}
	return string(rbuf.Bytes())
}
