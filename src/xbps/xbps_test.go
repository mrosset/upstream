package xbps

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
		err := tmpl.Save("./tmp/" + tmpl.Pkgname + ".json")
		if err != nil {
			t.Error(err)
		}
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

func TestRunDepend(t *testing.T) {
	d, err := GetRunDepends("telepathy-logger-devel")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("RUN ", d, "\n")
}

func TestBuildDepend(t *testing.T) {
	d, err := GetBuildDepends("gnome-shell")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("BUILD", d, "\n")
}

func TestDepends(t *testing.T) {
	for _, name := range []string{"xorg-server", "gnome-shell", "bash"} {
		fmt.Printf("checking depends for (%s)\n", name)
		rd, err := GetBuildDepends(name)
		sort.StringSlice(rd).Sort()
		if err != nil {
			t.Error(err)
		}
		req, err := ChkBuildDepends(rd)
		if err != nil {
			t.Error(err)
		}
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "\n**** (%s) Template Build Depends ****\n", name)
		printSlice(buf, rd)
		fmt.Fprintf(buf, "\n**** (%s) Actual Build Depend ****\n", name)
		printSlice(buf, req)

		fmt.Fprintf(buf, "\n%d template depends\n", len(rd))
		fmt.Fprintf(buf, "%2v actual depends\n", len(req))
		fmt.Fprintf(buf, "%2v removed\n", len(rd)-len(req))
		io.Copy(os.Stderr, buf)
	}
}


func printSlice(out io.Writer, slice []string) {
	for _, s := range slice {
		fmt.Fprintln(out, s)
	}
}


func printMap(out io.Writer, dmap map[string]bool) {
	for k, _ := range dmap {
		fmt.Fprintln(out, k)
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
