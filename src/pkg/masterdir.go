package xbps

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	. "github.com/str1ngs/go-ansi/color"
	"os/signal"
	"syscall"
)

var (
	hostPath = "/home/strings/pkg-cache"
	pkgDirs  = []string{
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
	log.Printf("%-10.10s %s", "mount", unExpand(target))
	fullcmd := "/usr/local/libexec/xbps-src-chroot-capmount"
	err = NewCommand("%s %s %s", fullcmd, source, target).Run()
	if err != nil {
		HandleError(err)
		return
	}
	return
}

func unbind(target string) (err os.Error) {
	log.Printf("%-10.10s %s", "umount", unExpand(target))
	fullcmd := "/usr/local/libexec/xbps-src-chroot-capumount"
	err = NewCommand("%s %s", fullcmd, target).Run()
	if err != nil {
		HandleError(err)
		return
	}
	return
}

func (this *MasterDir) signals() {
	for {
		select {
		case signal := <-signal.Incoming:
			switch signal.(os.UnixSignal) {
			case syscall.SIGTERM:
				log.Println(signal)
				this.UnMount()
				os.Exit(1)
			case syscall.SIGINT:
				log.Println(signal)
				this.UnMount()
				os.Exit(1)
			case syscall.SIGUSR1:
				log.Println(signal)
				this.UnMount()
				os.Exit(1)
			default:
				log.Print(signal)
			}
		}
	}
}
