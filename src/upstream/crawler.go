package main

import (
	"bytes"
	"fmt"
	"http"
	"io"
	"json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"xbps"
)

var (
	client = new(http.Client)
)

type Project struct {
	Id int
}

type Forge struct {
	Project Project
}

type Crawler struct {
	templates map[string]*xbps.Template
}

func NewCrawler(srcpath string) (*Crawler, os.Error) {
	ts, err := xbps.GetTemplates(srcpath)
	if err != nil {
		return nil, err
	}
	if len(ts) == 0 {
		return nil, fmt.Errorf("Expected greater then 0 templates go %v", len(ts))
	}
	return &Crawler{ts}, nil
}

func (c *Crawler) Start() os.Error {
	log.Println("staring", "crawl", "this could take some time...")

	efile, err := os.Create("upstream_error.log")
	if err != nil {
		log.Fatal(err)
	}
	defer efile.Close()

	nfile, err := os.Create("upstream_new.log")
	if err != nil {
		log.Fatal(err)
	}
	defer nfile.Close()

	for _, t := range c.templates {
		if t.Distfiles == "" {
			continue
		}
		cr, err := crawl(t)
		if err != nil {
			fmt.Fprintf(efile, "%s\n", err)
		}
		if cr.New {
			fmt.Fprintf(nfile, "%s\n", cr)
		}
	}
	return nil
}


func getParentUrl(t *xbps.Template) (string, string) {
	t.Distfiles = strings.Trim(t.Distfiles, " ")
	distfiles := strings.Split(t.Distfiles, " ")
	url, file := filepath.Split(distfiles[0])
	if url[:3] == "ftp" {
		url = "http" + url[3:]
	}
	return url, file
}

func Crawl(t *xbps.Template) (cr *CrawlResult, err os.Error) {
	cr, err = crawl(t)
	return cr, err
}

type CrawlResult struct {
	Name     string
	Vanilla  string
	Upstream string
	New      bool
}

func (this CrawlResult) String() string {
	return fmt.Sprintf("%-20.20s upstream %10.10s vanilla %10.10s %v",
		this.Name, this.Upstream, this.Vanilla, this.New)
}

func crawl(t *xbps.Template) (cr *CrawlResult, err os.Error) {
	cr = &CrawlResult{Name: t.Pkgname}
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()
	rawurl, file := getParentUrl(t)
	url, err := http.ParseURL(rawurl)
	if err != nil {
		return
	}
	// sourceforge sucks, make it suckless
	if url.Host == "downloads.sourceforge.net" {
		rawurl, err = getForgeUrl(t.Pkgname)
		if err != nil {
			return
		}
	}
	res, err := client.Get(rawurl)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return cr, fmt.Errorf("HTTP status %s", res.Status, rawurl)
	}
	rbuf := new(bytes.Buffer)
	_, err = io.Copy(rbuf, res.Body)
	if err != nil {
		return
	}
	//regex := `%s[a-z0-9\-\.]+(%s)`
	//fext := filepath.Ext(file)
	reg := regexp.MustCompile(fileToRegx(file))
	results := reg.FindAllString(string(rbuf.Bytes()), -1)
	if len(results) == 0 {
		return cr, fmt.Errorf("no distfiles found on %s", rawurl)
	}
	vregx := regxVersion(file)
	latest := findLatest(results, vregx)
	//if latest.Int > verInt(t.Version) {
	cr.Upstream = latest.String
	cr.Vanilla = t.Version
	cr.New = latest.Int > verInt(t.Version)
	return
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
		if vregx == regxVersion(file) {
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

func getForgeUrl(name string) (string, os.Error) {
	var (
		forgeJson = "http://sourceforge.net/api/project/name/%s/json"
		forgeRss  = "http://sourceforge.net/api/file/index/project-id/%v/mtime/desc/limit/20/rss"
	)
	res, err := client.Get(fmt.Sprintf(forgeJson, name))
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("HTTP status %s", res.Status)
	}
	defer res.Body.Close()
	forge := new(Forge)
	err = json.NewDecoder(res.Body).Decode(forge)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(forgeRss, forge.Project.Id), nil
}

func regxVersion(file string) string {
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
