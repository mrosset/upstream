package xbps

import (
	"bytes"
	"exec"
	"fmt"
	. "github.com/str1ngs/go-ansi/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	xbpsBin  = "fakeroot xbps-bin.static -c %s -Ayr %s %s %s"
	xbpsSrc  = "fakeroot xbps-src -C -m %s %s %s"
	mkIdx    = "xbps-src make-repoidx %s"
	binCache = "/home/strings/bincache"
	srcPath  = "/home/strings/github/vanilla/srcpkgs/"
	lfmt     = "%-10.10s %s"
	pkgfmt   = "%s-%s.x86_64.xbps"
	bufout   = new(bytes.Buffer)
)

func HandleError(err os.Error) {
	log.Print(err)
	log.Print(strings.Repeat("*", 80-len(log.Prefix())))
	io.Copy(os.Stderr, bufout)
	log.Print(strings.Repeat("*", 80-len(log.Prefix())))
}

func Seed(md *MasterDir) (err os.Error) {
	err = Install("base-chroot", md)
	if err != nil {
		return err
	}
	return
}

func Install(pack string, md *MasterDir) (err os.Error) {
	if IsInstalled(pack, md) {
		log.Printf(lfmt, "skip", pack)
		return
	}
	log.Printf(lfmt, "install", pack)
	err = NewCommand(xbpsBin, binCache, md.TargetPath, "install", pack).Run()
	if err != nil {
		return
	}
	return
}

func RmPackFile(pkg string) (err os.Error) {
	tmpl, err := FindTemplate(pkg, srcPath)
	if err != nil {
		return err
	}
	pkgname := fmt.Sprintf(pkgfmt, pkg, tmpl.Version)
	glob := fmt.Sprintf("%s/pkg-binpkgs/*/%s-*", hostPath, pkgname)
	files, err := filepath.Glob(glob)
	if err != nil {
		return
	}
	switch len(files) {
	case 0:
		log.Printf(lfmt, "no package", pkg)
		return
	case 1:
		log.Printf(lfmt, "remove", unExpand(files[0]))
		err = os.Remove(files[0])
		return err
	default:
		for _, f := range files {
			log.Printf(lfmt, "packages", unExpand(f))
		}
		return fmt.Errorf("expected one package to remove found %v", len(files))
	}
	panic("Should net reach here")
}

func Build(tmpl string, md *MasterDir) (err os.Error) {
	depends, err := GetDepends(tmpl)
	if err != nil {
		return err
	}
	log.Printf(lfmt, "depends", tmpl)
	for _, d := range depends {
		err = Install(d, md)
		if err != nil {
			return err
		}
	}
	log.Printf(lfmt, "build", tmpl)
	cmd := NewCommand(xbpsSrc, md.TargetPath, "install", tmpl)
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

func Package(tmpl string, md *MasterDir) (err os.Error) {
	if err = RmPackFile(tmpl); err != nil {
		return
	}
	log.Printf(lfmt, "package", tmpl)
	err = NewCommand(xbpsSrc, md.TargetPath, "build-pkg", tmpl).Run()
	if err != nil {
		return
	}

	log.Printf(lfmt, "index", tmpl)
	err = NewCommand(mkIdx, tmpl).Run()
	if err != nil {
		return
	}
	err = Clean(md)
	if err != nil {
		return
	}
	log.Printf(lfmt, "done", tmpl)
	return
}

func Clean(md *MasterDir) (err os.Error) {
	log.Printf(lfmt, "clean", md.TargetPath)
	if err = os.RemoveAll(md.TargetPath); err != nil {
		return
	}
	return
}

func IsInstalled(pack string, md *MasterDir) bool {
	cmd := NewCommand(xbpsBin, binCache, md.TargetPath, "show", TrimPack(pack))
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func TrimPack(pack string) string {
	return strings.Split(pack, ">")[0]
}

func NewCommand(format string, i ...interface{}) (cmd *exec.Cmd) {
	args := strings.Split(fmt.Sprintf(format, i...), " ")
	cmd = exec.Command(args[0], args[1:]...)
	bufout.Reset()
	bufout.WriteString(fmt.Sprintf("%s\n", Yellow(strings.Join(cmd.Args, " "))))
	cmd.Stdout = bufout
	cmd.Stderr = bufout
	return
}

func GetDepends(tmpl string) (depends []string, err os.Error) {
	b, err := Sh(fmt.Sprintf(printscript, tmpl))
	if err != nil {
		return
	}
	if len(b) == 0 {
		return
	}
	b = bytes.Trim(b, " ")
	depends = strings.Split(string(b), " ")
	return
}

func Sh(script string) (output []byte, err os.Error) {
	buf := new(bytes.Buffer)
	buf.WriteString(script)
	cmd := exec.Command("sh")
	cmd.Stdin = buf
	output, err = cmd.Output()
	if err != nil {
		os.Stderr.Write(output)
		return
	}
	return
}

var printscript = `
set -e 

. /usr/local/etc/xbps-src.conf
. /usr/local/share/xbps-src/shutils/init_funcs.sh
set_defvars
. /usr/local/share/xbps-src/shutils/tmpl_funcs.sh
. /usr/local/share/xbps-src/shutils/common_funcs.sh
. /usr/local/share/xbps-src/shutils/builddep_funcs.sh

setup_tmpl %s

echo -n $run_depends $build_depends
`
