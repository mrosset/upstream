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
	switch action {
	case "shell":
		tmpl, err := xbps.FindTemplate(target, *srcPath)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		io.Copy(os.Stdout, tmpl.ToSH())
	}
}
