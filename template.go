package main

import (
	"bytes"
	"exec"
	"fmt"
	"io"
	"json"
	"os"
	"path/filepath"
)

var shVars = []string{
	"pkgname",
	"version",
	"distfiles",
	"maintainer",
	"homepage",
	"license",
	"checksum",
}

type Template struct {
	Pkgname    string
	Version    string
	Distfiles  string
	ShortDesc  string "short_desc"
	Maintainer string
	Homepage   string
	License    string
	Checksum   string
	LongDesc   *json.RawMessage "long_desc"
	Path       string
}

func NewTemplate(file string) (*Template, os.Error) {
	os.Setenv("XBPS_SRCPKGDIR", *srcPath)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	buf := new(bytes.Buffer)
	buf.WriteString("Add_dependency(){ :\n }\n")
	io.Copy(buf, fd)
	buf.WriteString("echo {")
	for i, v := range shVars {
		line := fmt.Sprintf(`\"%s\":\"$%s\",`, v, v)
		if i == len(shVars)-1 {
			line = line[0 : len(line)-1]
		}
		buf.WriteString(line)
	}
	buf.WriteString("}")
	cmd := exec.Command("sh")
	cmd.Stdin = buf
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	template := new(Template)
	err = json.Unmarshal(output, template)
	if err != nil {
		return nil, err
	}
	template.Path = file
	return template, nil
}

func GetTemplates(srcpkgs string) (map[string]*Template, os.Error) {
	var (
		errors    = 0
		linked    = 0
		ok        = 0
		templates = map[string]*Template{}
	)
	files, err := filepath.Glob(srcpkgs + "/*/template")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		d, _ := filepath.Split(file)
		d = filepath.Clean(d)
		di, err := os.Stat(d)
		if err != nil {
			return nil, err
		}
		if di.FollowedSymlink {
			linked++
			continue
		}
		t, err := NewTemplate(file)
		if err != nil {
			errors++
			continue
		}
		ok++
		if t != nil {
			templates[t.Pkgname] = t
		}
	}
	/*
		total := errors + linked + ok
		fmt.Println("errors",errors)
		fmt.Println("linked",linked)
		fmt.Println("ok",ok)
		fmt.Println("total",total)
	*/
	return templates, nil
}