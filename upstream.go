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
	"strings"
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
	isHome     = flag.Bool("home", false, "open package homepage in browser")
	isBroken   = flag.Bool("b", false, "print out broken upstream versions")
	isCheck    = flag.Bool("c", false, "check upstream version against our versions")
	isSync     = flag.Bool("s", false, "sync oswatershed and srpkgs cache")
	isLongDesc = flag.Bool("ld", false, "get long description from debian packages")
	srcPath    = flag.String("path", "/home/strings/github/vanilla/srcpkgs/", "path to srcpkgs")
	browser    = flag.String("browser", "chromium", "broswer to use")
)

type Distro struct {
	Version string
}

type Package struct {
	Name    string "package"
	Latest  string
	Vanilla string
	Distros []Distro
	Check   bool
}

func init() {
	log.SetPrefix("")
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
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
		err := check()
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
	}
	pack, err := latest(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pack.Latest)
}

func home(pack string) os.Error {
	template := filepath.Join(*srcPath, pack, "template")
	home, err := getVar(template, "homepage")
	if err != nil {
		return err
	}
	fmt.Println("opening", home)
	err = exec.Command(*browser, home).Run()
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
			fmt.Printf("%v\n", p.Name)
		}
	}
	return nil

}

func check() os.Error {
	cache, err := loadCache()
	if err != nil {
		return err
	}
	fmt.Printf("%-20.20s %-20.20s %-20.20s %-20.20s\n", "name", "vanilla", "upstream", "archlinux")
	for _, p := range *cache {
		if p.Latest != p.Vanilla && p.Check {
			fmt.Printf("%-20.20s %-20.20s %-20.20s %-20.20s\n", p.Name, p.Vanilla, p.Latest, p.Distros[ARCHLINUX].Version)
		}
	}
	fmt.Printf("%-20.20s %-20.20s %-20.20s %-20.20s\n", "name", "vanilla", "upstream", "archlinux")
	return nil
}

func sync() os.Error {
	cache := map[string]*Package{}
	templates, err := filepath.Glob(*srcPath + "/*/template")
	if err != nil {
		log.Fatal(err)
	}
	for i, template := range templates {
		_, err := os.Stat(template)
		if err != nil {
			return err
		}
		dir, _ := filepath.Split(template)
		pname := filepath.Base(dir)
		ver, err := getVar(template, "version")
		if err != nil {
			return err
		}
		pack, err := latest(pname)
		if err != nil {
			p := new(Package)
			p.Name = pname
			p.Check = false
			p.Vanilla = ver
			fmt.Printf("%04.0v %-20.20s %s\n", i, pname, "error")
			cache[pname] = p
			continue
		}
		pack.Vanilla = ver
		pack.Check = true
		cache[pname] = pack
		fmt.Printf("%04.0v %-20.20s %-10.10s\n", i, pname, pack.Latest)
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

func getVar(template string, shvar string) (string, os.Error) {
	os.Setenv("XBPS_SRCPKGDIR", *srcPath)
	fd, err := os.Open(template)
	if err != nil {
		return "", err
	}
	defer fd.Close()
	buf := new(bytes.Buffer)
	buf.WriteString("Add_dependency(){ :\n }\n")
	io.Copy(buf, fd)
	buf.WriteString("echo $" + shvar)
	cmd := exec.Command("sh")
	cmd.Stdin = buf
	output, err := cmd.CombinedOutput()
	if err != nil {
		println(template)
		log.Println(err, string(output))
		return "", err
	}
	svar := strings.Replace(string(output), "\n", "", -1)
	if svar == "" {
		return "", fmt.Errorf("%s not found in %s", shvar, template)
	}
	return svar, err
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
