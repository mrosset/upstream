package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"http"
	"io"
	"json"
	"log"
	"os"
	"path/filepath"
)

var packages = []string{
	"bash",
	"grep",
	"kernel",
	"curl",
	"libcurl",
	"rtorrent",
	"libX11",
	"failthis",
}

const (
	watershed = "http://api.oswatershed.org/api/0.1/package.json?package=%s&cb=go"
	debian    = "http://packages.debian.org/%s/%s"
	drelease  = "wheezy"
)

var (
	isLongDesc = flag.Bool("ld", false, "get long description from debian packages")
	isTest     = flag.Bool("t", false, "run tests")
	srcPath    = flag.String("path", "/home/strings/github/vanilla/srcpkgs/", "path to srcpkgs")
)

type Distro struct {
	Version string
}

type Package struct {
	Name    string "package"
	Latest  string
	Distros []Distro
}

func init() {
	log.SetPrefix("watch: ")
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
	if *isTest {
		test()
		return
	}
	if len(flag.Args()) != 1 {
		flag.Usage()
		return
	}
	if *isLongDesc {
		err := longDesc(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	pack, err := latest(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pack.Latest)
}

func test() {
	packages, err := filepath.Glob(*srcPath + "/*")
	if err != nil {
		log.Fatal(err)
	}
	for i, dir := range packages {
		if i >= 20 {
			break
		}
		_, dir := filepath.Split(dir)
		pack, err := latest(dir)
		if err != nil {
			fmt.Printf("%04.0v %-20.20s %s\n", i, dir, "error")
			continue
		}
		fmt.Printf("%04.0v %-20.20s %-10.10s arch %s\n", i, dir, pack.Latest, pack.Distros[0].Version)
	}
}

func longDesc(name string) os.Error {
	client := new(http.Client)
	url := fmt.Sprintf(debian, drelease, name)
	res, err := client.Do(request(url))
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

func latest(name string) (*Package, os.Error) {
	client := new(http.Client)
	url := fmt.Sprintf(watershed, name)
	res, err := client.Do(request(url))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Http GET error %s", res.Status)
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)
	p := new(Package)
	b := buf.Bytes()
	err = json.Unmarshal(b[3:len(b)-1], p)
	if err != nil {
		fmt.Println(err)
		return p, err
	}
	return p, nil
}

func request(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.TransferEncoding = []string{"chunked"}
	//FIXME: Debian doesnt always return compressed
	//req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Connection", "keep-alive")
	return req
}
