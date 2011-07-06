package xbps

import (
	"testing"
)


var packs = []string{"ncdu"}

func TestTest(t *testing.T) {
	md, err := NewMasterDir("masterdir")
	if err != nil {
		HandleError(err)
		t.Fatal(err)
	}
	if md == nil {
		t.Fatal("MasterDir is nil")
	}
	err = md.Mount()
	if err != nil {
		HandleError(err)
		md.UnMount()
		t.Fatal(err)
	}
	err = Seed(md)
	if err != nil {
		HandleError(err)
		t.Fatal(err)
	}
	for _, p := range packs {
		err = Build(p, md)
		if err != nil {
			HandleError(err)
			md.UnMount()
			t.Fatal(err)
		}
		err = Package(p, md)
		if err != nil {
			HandleError(err)
			md.UnMount()
			t.Fatal(err)
		}
	}
}
