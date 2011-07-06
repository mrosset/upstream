package xbps

import (
	"exec"
	"fmt"
	"log"
	"os"
	"path/filepath"
	. "github.com/str1ngs/go-ansi/color"
)

var (
	targetPath = "/home/strings/masters"
	hostPath   = "/home/strings/pkg-cache"
	pkgDirs    = []string{
		"pkg-binpkgs",
		"pkg-srcdistdir",
	}
)

func init() {
	log.SetPrefix(fmt.Sprintf("%s", Blue("masterdir: ")))
	log.SetFlags(0)
}

type MasterDir struct {
	Name       string
	HostPath   string
	PkgDirs    []string
	TargetPath string
}

func NewMasterDir(name string) (md *MasterDir, err os.Error) {
	md = &MasterDir{
		Name:       name,
		TargetPath: filepath.Join(targetPath, name),
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

func (this *MasterDir) Mount() (err os.Error) {
	for _, dir := range this.PkgDirs {
		host := filepath.Join(this.HostPath, dir)
		target := filepath.Join(this.TargetPath, dir)
		err = bind(host, target)
		if err != nil {
			return
		}
	}
	return
}

func (this *MasterDir) UnMount() (err os.Error) {
	for _, dir := range this.PkgDirs {
		target := filepath.Join(this.TargetPath, dir)
		err = unbind(target)
		if err != nil {
			return
		}
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

func bind(source string, target string) (err os.Error) {
	log.Printf("%-10.10s %-40.40s", "mount", target)
	fullcmd := "/usr/local/libexec/xbps-src-chroot-capmount"
	output, err := exec.Command(fullcmd, source, target).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return
	}
	return
}

func unbind(target string) (err os.Error) {
	log.Printf("%-10.10s %-40.40s", "umount", target)
	fullcmd := "/usr/local/libexec/xbps-src-chroot-capumount"
	output, err := exec.Command(fullcmd, target).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return
	}
	return
}
