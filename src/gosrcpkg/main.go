package main

import (
	"flag"
	"io"
	"log"
	"os"
	"xbps"
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		os.Exit(1)
	}
	action := flag.Arg(0)
	target := flag.Arg(1)

	tmpl, err := xbps.FindTemplate(target, xbps.SRCPKGDIR)
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
	case "chkdepends":
		_, err := xbps.ChkDupDepends(target)
		if err != nil {
			log.Fatal(err)
		}
	}
}
