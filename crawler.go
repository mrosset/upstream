package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	nrange := 0
	for _, t := range c.templates {
		nrange++
		if t.Distfiles == "" {
			continue
		}
		c.crawl(t)
	}
	return nil
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
	url, file := getParentUrl(t)
	println(url, file)
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
