package main

import (
	"bytes"
	"flag"
	"fmt"
	"http"
	"io"
	"json"
	"log"
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

const url = "http://api.oswatershed.org/api/0.1/package.json?package=%s&cb=go"

var (
	client = new(http.Client)
	isTest = flag.Bool("t", false, "run tests")
)

type Package struct {
	Name    string "package"
	Latest  string
	Distros *json.RawMessage
}

func main() {
	flag.Parse()
	if *isTest {
		test()
	}
}

func test() {
	fmt.Printf("%-20.20s %s\n", "package", "version")
	fmt.Printf("%-20.20s %s\n", "-------", "-------")
	c := make(chan int)
	for _, p := range packages {
		go latest(p, c)
	}
	for i := 0; i < len(packages); i++ {
		<-c
	}
}

func latest(name string, c chan int) {
	res, err := client.Do(request(name))
	defer func() { c <- 1 }()
	if err != nil {
		log.Println(err)
		return
	}
	if res.StatusCode != 200 {
		fmt.Printf("%-20.20s %s\n", name, "error")
		return
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)
	p := new(Package)
	b := buf.Bytes()
	err = json.Unmarshal(b[3:len(b)-1], p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%-20.20s %s\n", p.Name, p.Latest)
	return
}

func request(name string) *http.Request {
	req, err := http.NewRequest("GET", fmt.Sprintf(url, name), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.URL, err = http.ParseURL(fmt.Sprintf(url, name))
	if err != nil {
		log.Fatal(err)
	}
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.TransferEncoding = []string{"chunked"}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Connection", "keep-alive")
	return req
}
