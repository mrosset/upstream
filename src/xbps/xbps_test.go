package xbps

import (
	"bytes"
	"io"
	"os"
	"testing"
)

var (
	packs  = []string{"ncdu"}
	master *MasterDir
	spath  = "/home/strings/github/vanilla/srcpkgs"
	dprint = false
)

func TestSerializeAll(t *testing.T) {
	tmpls, err := GetTemplates(spath)
	if err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	for _, tmpl := range tmpls {
		r := tmpl.ToSH()
		io.Copy(buf, r)
		r, err = tmpl.ToJson()
		if err != nil {
			t.Error(err)
		}
		io.Copy(buf, r)
		if dprint {
			io.Copy(os.Stderr, buf)
		}
	}
}

func TestSh(t *testing.T) {
	_, err := FindTemplate("bash", spath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJson(t *testing.T) {
	tmpl, err := FindTemplate("bash", spath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmpl.ToJson()
	if err != nil {
		t.Fatal(err)
	}
}

/*
func TestPrint(t *testing.T) {
	tmpl, err := FindTemplate("bash", spath)
	if err != nil {
		t.Fatal(err)
	}
	in := new(bytes.Buffer)
	err = json.NewEncoder(in).Encode(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	out := new(bytes.Buffer)
	err = json.Indent(out, in.Bytes(), "", "")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(os.Stderr, out)
}

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
