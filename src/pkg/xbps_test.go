package xbps

import (
	"os"
	"testing"
)

var (
	packs  = []string{"ncdu", "lsof", "pulseaudio"}
	master *MasterDir
)

func TestNewMaster(t *testing.T) {
	var err os.Error
	master, err = NewMasterDir("masterdir", "./masters")
	if err != nil {
		HandleError(err)
		t.Fatal(err)
	}
}

func TestMount(t *testing.T) {
	err := master.Mount()
	if err != nil {
		HandleError(err)
		master.UnMount()
		t.Fatal()
	}
}

func TestSeed(t *testing.T) {
	err := Seed(master)
	if err != nil {
		HandleError(err)
		master.UnMount()
		t.Fatal()
	}
}

func TestBuild(t *testing.T) {
	for _, pkg := range packs {
		err := Build(pkg, master)
		if err != nil {
			HandleError(err)
			master.UnMount()
			t.Fatal(err)
		}
	}
}

func TestPackage(t *testing.T) {
	for _, pkg := range packs {
		err := Package(pkg, master)
		if err != nil {
			HandleError(err)
			master.UnMount()
			t.Fatal(err)
		}
	}
	err := Clean(master)
	if err != nil {
		t.Fatal(err)
	}
}
