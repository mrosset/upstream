package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"xbps"
)

func main() {
	flag.Parse()
	var target string
	switch len(flag.Args()) {
	case 0:
		wd, _ := os.Getwd()
		target = filepath.Base(wd)
	case 1:
		target = flag.Arg(0)
	}
	md, err := xbps.NewMasterDir("masterdir", "/home/strings/masters/")
	if err != nil {
		log.Fatal(err)
	}
	if err = xbps.Seed(md); err != nil {
		xbps.HandleError(err)
		os.Exit(1)
	}
	if err = md.Mount(); err != nil {
		log.Fatal(err)
	}
	if err = xbps.Build(target, md); err != nil {
		xbps.HandleError(err)
		md.UnMount()
		os.Exit(1)
	}
	if err = xbps.Package(target, md); err != nil {
		xbps.HandleError(err)
		md.UnMount()
	}
	return
}
