package main

import (
	"flag"
	"io"
	"log"
	"os"
	"xbps"
)

var (
	srcPath = flag.String("path", "/home/strings/github/vanilla/srcpkgs/", "path to srcpkgs")
)

func init() {
	log.SetPrefix("")
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		os.Exit(1)
	}
	action := flag.Arg(0)
	target := flag.Arg(1)

	tmpl, err := xbps.FindTemplate(target, *srcPath)
	if err != nil {
		log.Fatal(err)
	}

	switch action {
	case "shell":
		io.Copy(os.Stdout, tmpl.ToSH())
	case "json":
		js, err := tmpl.ToJson()
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, js)
	}
}
