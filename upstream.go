package main

import (
	"bytes"
	"exec"
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

const (
	watershed = "http://api.oswatershed.org/api/0.1/package.json?package=%s&cb=go"
	debian    = "http://packages.debian.org/%s/%s"
	drelease  = "wheezy"
)

const (
	ARCHLINUX = iota
)

var (
	isTest          = flag.Bool("test", false, "run tests")
	isCheckTemplate = flag.Bool("ct", false, "check templates *currently* only checks homepage, license")
	isHome          = flag.Bool("home", false, "open package homepage in browser")
	isBroken        = flag.Bool("b", false, "print out broken upstream versions")
	isCheck         = flag.Bool("c", false, "check upstream version against our versions")
	isSync          = flag.Bool("s", false, "sync oswatershed and srpkgs cache")
	isLongDesc      = flag.Bool("ld", false, "get long description from debian packages")
	srcPath         = flag.String("path", "/home/strings/github/vanilla/srcpkgs/", "path to srcpkgs")
	browser         = flag.String("browser", "chromium", "browser to use")
)

type Distro struct {
	Version string
}

type Package struct {
	Package string
	Latest  string
	Vanilla string
	Distros []Distro
	Check   bool
}

func init() {
	log.SetFlags(0)
}

func main() {
	flag.Parse()
	lfile, err := os.Create("upstream_error.log")
	if err != nil {
		log.Fatal(err)
	}
	defer lfile.Close()
	log.SetOutput(lfile)

	if *isTest {
		test()
		return
	}
	if *isCheckTemplate {
		err := checkTemplates()
		if err != nil {
			log.Print(err)
		}
		return
	}
	if *isHome {
		pack := flag.Arg(0)
		if pack == "" {
			flag.Usage()
			return
		}
		err := home(pack)
		if err != nil {
			log.Print(err)
		}
		return
	}
	if *isSync {
		err := sync()
		if err != nil {
			log.Print(err)
		}
		return
	}
	if *isBroken {
		err := broken()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if *isCheck {
		err := checkVersions()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if *isLongDesc {
		err := longDesc(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if len(flag.Args()) != 1 {
		flag.Usage()
		return
	}
	pack, err := latest(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pack.Latest)
}

func test() {
	ps := []string{
		"bash",
		"wget",
		"chromium",
		"tar",
	}
	c, err := NewCrawler(*srcPath)
	checkFatal(err)
	c.Start()
	fmt.Println(ps)
}

func checkFatal(err os.Error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkTemplates() os.Error {
	templates, err := GetTemplates(*srcPath)
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

func home(pack string) os.Error {
	file := filepath.Join(*srcPath, pack, "template")
	t, err := NewTemplate(file)
	if err != nil {
		return err
	}
	if t.Homepage == "" {
		return fmt.Errorf("%s has no homepage", pack)
	}
	fmt.Println("opening", t.Homepage)
	err = exec.Command(*browser, t.Homepage).Run()
	if err != nil {
		return err
	}

	return nil
}

func broken() os.Error {
	cache, err := loadCache()
	if err != nil {
		return err
	}
	for _, p := range *cache {
		if !p.Check {
			fmt.Printf("%v\n", p.Package)
		}
	}
	return nil

}

func checkVersions() os.Error {
	maperror := 0
	cache, err := loadCache()
	if err != nil {
		return err
	}
	ts, err := GetTemplates(*srcPath)
	if err != nil {
		return err
	}
	fmt.Printf("%-20.20s %-20.20s %-20.20s %-20.20s\n", "name", "vanilla", "upstream", "archlinux")
	for _, p := range *cache {
		if ts[p.Package] == nil {
			maperror++
			continue
		}
		if p.Latest != ts[p.Package].Version && p.Check {
			fmt.Printf("%-20.20s %-20.20s %-20.20s %-20.20s\n", p.Package, ts[p.Package].Version, p.Latest, p.Distros[ARCHLINUX].Version)
		}
	}
	fmt.Printf("%-20.20s %-20.20s %-20.20s %-20.20s\n", "name", "vanilla", "upstream", "archlinux")
	fmt.Println(maperror, "map errors")
	return nil
}

func sync() os.Error {
	cache := map[string]*Package{}
	templates, err := GetTemplates(*srcPath)
	checkFatal(err)
	nrange := 0
	for _, t := range templates {
		nrange++
		pack, err := latest(t.Pkgname)
		if err != nil {
			p := new(Package)
			p.Package = t.Pkgname
			p.Check = false
			cache[t.Pkgname] = p
			fmt.Printf("%04.0v %-20.20s %s\n", nrange, t.Pkgname, "error")
			continue
		}
		pack.Check = true
		cache[t.Pkgname] = pack
		fmt.Printf("%04.0v %-20.20s %-10.10s\n", nrange, t.Pkgname, pack.Latest)
	}
	f, err := os.Create(*srcPath + "/upstream.json")
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = json.Indent(buf, b, "", "\t")
	if err != nil {
		return err
	}
	_, err = io.Copy(f, buf)
	if err != nil {
		return err
	}
	return nil
}

func loadCache() (*map[string]*Package, os.Error) {
	cache := &map[string]*Package{}
	fd, err := os.Open(*srcPath + "upstream.json")
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	err = json.NewDecoder(fd).Decode(cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
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
