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
	"sort"
	"strings"
)

var (
	xbpsBin  = "xbps-bin.static -c %s -Ayr %s %s %s"
	xbpsSrc  = "xbps-src -C -m %s %s %s"
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
	panic("Should not reach here")
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
	rd, err := GetRunDepends(tmpl)
	if err != nil {
		return
	}
	bd, err := GetBuildDepends(tmpl)
	if err != nil {
		return
	}
	depends = append(rd, bd...)
	return
}

func GetRunDepends(tmpl string) (depends []string, err os.Error) {
	b, err := Sh([]string{"run", tmpl})
	if err != nil {
		return
	}
	if len(b) == 0 {
		return
	}
	if b[0] == ' ' {
		b = b[1:]
	}
	depends = strings.Split(string(b), " ")
	trimOper(depends)
	sort.StringSlice(depends).Sort()
	return
}

func GetBuildDepends(tmpl string) (depends []string, err os.Error) {
	b, err := Sh([]string{"build", tmpl})
	if err != nil {
		return
	}
	if len(b) == 0 {
		return
	}
	if b[0] == ' ' {
		b = b[1:]
	}
	depends = strings.Split(string(b), " ")
	trimOper(depends)
	return
}

// Checks each build dependency and makes sure it is unique
func ChkBuildDepends(depends []string) (required []string, err os.Error) {
	mreq := make(map[string]bool)
	var (
		visited = make(map[string]string)
	)
	// walk each depend
	for _, d := range depends {
		if !strings.HasSuffix(d, "-devel") {
			continue
		}
		c, err := GetRunDepends(d)
		if err != nil {
			return
		}

		// walk each sub depend and mark it as visited
		for _, sc := range c {
			visited[sc] = d
			// if this sub depend is in required remove it
			if mreq[sc] {
				fmt.Printf("%-20.20s provided by %s\n", sc, visited[sc])
				mreq[sc] = false, false
			}
		}

		// if we visited this depend before it is not required
		_, exist := visited[d]
		if !exist {
			mreq[d] = true
		} else {
			fmt.Printf("%-20.20s provided by %s\n", d, visited[d])
		}
	}
	for r, _ := range mreq {
		required = append(required, r)
	}
	sort.StringSlice(required).Sort()
	return
}

func trimOper(depends []string) {
	for i, d := range depends {
		kv := strings.Split(d, ">=")
		depends[i] = kv[0]
	}
}

func Sh(args []string) (output []byte, err os.Error) {
	helper := "/usr/local/libexec/xbps-src-getdeps-helper"
	cmd := exec.Command(helper, args...)
	output, err = cmd.Output()
	if err != nil {
		os.Stderr.Write(output)
		return
	}
	return
}

/*
`
set -e 

. /usr/local/etc/xbps-src.conf
. /usr/local/share/xbps-src/shutils/init_funcs.sh
set_defvars
. /usr/local/share/xbps-src/shutils/tmpl_funcs.sh
. /usr/local/share/xbps-src/shutils/common_funcs.sh
. /usr/local/share/xbps-src/shutils/builddep_funcs.sh

setup_tmpl %s

echo -n $%s_depends
`
*/
