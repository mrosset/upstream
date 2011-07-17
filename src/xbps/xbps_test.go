package xbps

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)


var (
	master *MasterDir
	// set your srcpkgs
	spath  = "/home/strings/github/vanilla/srcpkgs"
	dprint = false
)

func init() {
	_, err := os.Stat(spath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("please set spath to your srcpkgs locacation in xbps_test.go")
		os.Exit(0)
	}
}

var tpacks = []string{"ncdu"}

func TestTemplate(t *testing.T) {
	for _, pkg := range tpacks {
		tmpl, err := FindTemplate(pkg, spath)
		if err != nil {
			t.Error(err)
		}
		if tmpl.Pkgname != pkg {
			t.Error(err)
		}
	}
}


func TestSerializeAll(t *testing.T) {
	tmpls, err := GetTemplates(spath)
	if err != nil {
		t.Error(err)
	}
	buf := new(bytes.Buffer)
	for _, tmpl := range tmpls {
		r := tmpl.ToSH()
		io.Copy(buf, r)
		r, err = tmpl.ToJson()
		if err != nil {
			t.Errorf("%s %s", tmpl.Pkgname, err)
		}
		io.Copy(buf, r)
		if dprint {
			io.Copy(os.Stderr, buf)
		}
		/*
			err := tmpl.Save("./tmp/" + tmpl.Pkgname + ".json")
			if err != nil {
				t.Error(err)
			}
		*/
	}
}


func TestLoad(t *testing.T) {
	files, err := filepath.Glob("./tmp/*.json")
	if err != nil {
		t.Error(err)
	}
	for _, f := range files {
		_, err := LoadJson(f)
		if err != nil {
			t.Error(err)
		}
	}
}

func DisableTrims(t *testing.T) {
	ts, err := GetTemplates(spath)
	if err != nil {
		t.Error(err)
	}
	for _, tm := range ts {
		_, err := GetAllDepends(tm.Pkgname)
		if err != nil {
			t.Errorf("%s: %s", tm.Pkgname, err)
		}
	}
}

type DepTest struct {
	name     string
	expected string
}


var deptests = []DepTest{
	{"fontconfig", "pkg-config freetype-devel expat-devel"},
	{"libxslt-devel", "libgcrypt-devel libxml2-devel python-devel libxslt"},
}


func TestDepends(t *testing.T) {
	for _, test := range deptests {
		depends, err := ChkDupDepends(test.name)
		if err != nil {
			t.Error(err)
		}
		result := strings.Join(depends, " ")
		if result != test.expected {
			t.Errorf("%s: expected %s got %s", test.name, test.expected, result)
		}
	}
}
