package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"http"
	"io"
	"log"
	"os"
	"xbps"
)

func usage() {
	fmt.Fprint(os.Stderr, "usage: upstream <flag> <action> <targets>\n\n")
	fmt.Fprint(os.Stderr, "flags\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
	fmt.Fprint(os.Stderr, "actions\n")
	fmt.Fprintf(os.Stderr, "%-17.17s %s\n", "check <targets>", "check each target for newer versions upstream")
	fmt.Fprintf(os.Stderr, "%-17.17s %s\n", "crawl ", "crawl all templates and log new packages to ./upstream_new.log")
	os.Exit(2)
}

const (
	debian   = "http://packages.debian.org/%s/%s"
	drelease = "wheezy"
)

func init() {
	log.SetFlags(0)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	action := flag.Args()[0]
	targets := flag.Args()[1:]
	switch action {
	case "check":
		Check(targets)
		return
	case "crawl":
		docrawl()
		return
	default:
		flag.Usage()
	}
}


func Check(names []string) {
	for _, n := range names {
		t, err := xbps.FindTemplate(n, xbps.SRCPKGDIR)
		if err != nil {
			fmt.Printf("%s: %s\n", n, err)
			continue
		}
		cr, err := Crawl(t)
		fmt.Println(cr)
	}
}


func docrawl() {
	ps := []string{
		"xstow",
	}
	_ = ps
	c, err := NewCrawler(xbps.SRCPKGDIR)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	/*
			for _, p := range ps {
				c.Crawl(p)
		}
	*/
	c.Start()
}

func checkFatal(err os.Error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkTemplates() os.Error {
	templates, err := xbps.GetTemplates(xbps.SRCPKGDIR)
	if err != nil {
		return err
	}
	for _, t := range templates {
		if t.License == "" || t.Homepage == "" {
			fmt.Println(t.Pkgname)
		}
	}
	return nil
}

func longDesc(name string) os.Error {
	client := new(http.Client)
	url := fmt.Sprintf(debian, drelease, name)
	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Http GET error %s", res.Status)
	}
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return err
	}
	doc, err := html.Parse(buf)
	if err != nil {
		return err
	}
	//TODO: move this to a proper function
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			if len(n.Attr) > 0 {
				if n.Attr[0].Val == "pdesc" {
					fmt.Printf("%s\n\n", n.Child[1].Child[0].Data)
					fmt.Printf("*long desc*\n%s", n.Child[1].Child[2].Child[0].Data)
				}
			}
		}
		for _, c := range n.Child {
			f(c)
		}
	}
	f(doc)
	return nil
}
