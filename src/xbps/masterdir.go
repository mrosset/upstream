package xbps

import (
	"log"
	"os"
	"path/filepath"
)

var (
	hostPath = "/home/strings/host"
	pkgDirs  = []string{
		"pkg-binpkgs",
		"pkg-srcdistdir",
	}
)

func init() {
	argv0 := os.Args[0]
	log.SetPrefix(argv0 + ": ")
	log.SetFlags(0)
}

type MasterDir struct {
	Name       string
	HostPath   string
	PkgDirs    []string
	TargetPath string
}

func NewMasterDir(name, mpath string) (md *MasterDir, err os.Error) {
	fullpath, err := filepath.Abs(mpath)
	if err != nil {
		return
	}
	md = &MasterDir{
		Name:       name,
		TargetPath: filepath.Join(fullpath, name),
		HostPath:   hostPath,
		PkgDirs:    pkgDirs,
	}
	if err = md.mkPkgDirs(md.HostPath); err != nil {
		return
	}
	if err = md.mkPkgDirs(md.TargetPath); err != nil {
		return
	}
	return
}

func (this *MasterDir) mkPkgDirs(parent string) (err os.Error) {
	for _, dir := range this.PkgDirs {
		fullpath := filepath.Join(parent, dir)
		err = os.MkdirAll(fullpath, 0755)
		if err != nil {
			return
		}
	}
	return
}
