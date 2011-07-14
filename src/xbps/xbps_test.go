package xbps

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

var (
	tpacks = []string{"ncdu"}
	master *MasterDir
	spath  = "/home/strings/github/vanilla/srcpkgs"
	dprint = false
)

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
	}
}

func TestDepends(t *testing.T) {
	name := "gnome-bluetooth-devel"
	rd, err := GetRunDepends(name)
	if err != nil {
		t.Error(err)
	}
	req, err := ChkRunDepends(rd)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("\n**** (%s) Template Run Depends ****\n", name)
	printSlice(rd)
	fmt.Printf("\n**** (%s) Actual Run Depend ****\n", name)
	printMap(req)

	fmt.Printf("\n%d template depends\n", len(rd))
	fmt.Printf("%2v actual depends\n", len(req))
	fmt.Printf("%2v removed\n", len(rd)-len(req))
}


func printSlice(slice []string) {
	for _, s := range slice {
		fmt.Println(s)
	}
}

func printMap(dmap map[string]bool) {
	for k, _ := range dmap {
		fmt.Println(k)
	}
}
/*
func TestNewMaster(t *testing.T) {
	var err os.Error
	master, err = NewMasterDir("masterdir", "/home/strings/masters")
	if err != nil {
		HandleError(err)
		t.Fatal(err)
	}
}

func TestSeed(t *testing.T) {
	err := Seed(master)
	if err != nil {
		HandleError(err)
		t.Fatal()
	}
}

func TestBuild(t *testing.T) {
	for _, pkg := range packs {
		err := Build(pkg, master)
		if err != nil {
			HandleError(err)
			t.Fatal(err)
		}
	}
}

func TestPackage(t *testing.T) {
	for _, pkg := range packs {
		err := Package(pkg, master)
		if err != nil {
			HandleError(err)
			t.Fatal(err)
		}
	}
	err := Clean(master)
	if err != nil {
		t.Fatal(err)
	}
}
*/
