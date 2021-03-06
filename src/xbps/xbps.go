package xbps

import (
	"bytes"
	"exec"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	xbpsBin   = "xbps-bin.static -c %s -Ayr %s %s %s"
	xbpsSrc   = "xbps-src -C -m %s %s %s"
	mkIdx     = "xbps-src make-repoidx %s"
	binCache  = "/home/strings/bincache"
	lfmt      = "%-10.10s %s"
	SRCPKGDIR string
	pkgfmt    = "%s-%s.x86_64.xbps"
	bufout    = new(bytes.Buffer)
)

func init() {
	argv0 := os.Args[0]
	log.SetPrefix(argv0 + ": ")
	log.SetFlags(log.Lshortfile)

	SRCPKGDIR = os.Getenv("XBPS_SRCPKGDIR")
	if SRCPKGDIR == "" {
		srcerror()
	}
	_, err := os.Stat(SRCPKGDIR)
	if err != nil {
		srcerror()
	}
}

func srcerror() {
	log.Fatal("you must set XBPS_SRCPKGDIR enviroment variable to use xbps-go")
}

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
	tmpl, err := FindTemplate(pkg, SRCPKGDIR)
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
	depends, err := GetAllDepends(tmpl)
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
	cmd := NewCommand(xbpsBin, binCache, md.TargetPath, "show", TrimOp(pack))
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

// Creates a new exec.Cmd from a Printf style format
func NewCommand(format string, i ...interface{}) (cmd *exec.Cmd) {
	args := strings.Split(fmt.Sprintf(format, i...), " ")
	cmd = exec.Command(args[0], args[1:]...)
	bufout.Reset()
	bufout.WriteString(fmt.Sprintf("%s\n", strings.Join(cmd.Args, " ")))
	cmd.Stdout = bufout
	cmd.Stderr = bufout
	return
}

// Get all dependancies for a template
func GetAllDepends(tmpl string) (depends []string, err os.Error) {
	rd, err := GetDepends("run", tmpl)
	if err != nil {
		return
	}
	bd, err := GetDepends("build", tmpl)
	if err != nil {
		return
	}
	depends = append(rd, bd...)
	return
}

// Get run/build depends for a template
func GetDepends(kind, tmpl string) (depends []string, err os.Error) {
	b, err := Sh([]string{kind, tmpl})
	if err != nil {
		return
	}
	if len(b) == 0 {
		return
	}
	b = bytes.Trim(b, " ")
	b = bytes.Trim(b, "\n")
	depends = strings.Split(string(b), " ")
	TrimOps(depends)
	return
}

func ChkDupDepends(name string) (required []string, err os.Error) {
	fmt.Printf("**** Checking Depends For (%s) ****\n", name)
	var (
		mreq    = make(map[string]bool)
		visited = make(map[string]string)
		depends = []string{}
	)
	if isSubTmpl(name) {
		depends, err = GetDepends("run", name)
	} else {
		depends, err = GetDepends("build", name)
	}
	if err != nil {
		return
	}
	// walk each depend
	for _, d := range depends {
		var c []string
		c, err = GetDepends("run", d)
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
	fmt.Printf("\n*** The following lines must be removed from (%s) *** \n\n", name)
	for _, d := range depends {
		if !sContains(required, d) {
			fmt.Println("Add_dependancy build", d)
		}
	}
	fmt.Println()
	return
}

// Remove all version chars from package name
func TrimOp(pack string) string {
	var (
		name    []int
		version []int
	)
	var prefix = true
	for _, c := range pack {
		switch c {
		case '<', '>', '=', '[', ']', '*', '.':
			version = append(version, c)
			prefix = false
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if prefix {
				name = append(name, c)
			} else {
				version = append(version, c)
			}
		default:
			name = append(name, c)
		}
	}
	return string(name)
}

func TrimOps(depends []string) {
	for i, d := range depends {
		depends[i] = TrimOp(d)
	}
}

func isSubTmpl(tmpl string) bool {
	file := fmt.Sprintf("%s/%s/%s.template", SRCPKGDIR, tmpl, tmpl)
	return fileExists(file)
}

func Sh(args []string) (output []byte, err os.Error) {
	helper := "/usr/local/libexec/getdeps-helper"
	cmd := exec.Command(helper, args...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s %s", helper, err, output)
	}
	return
}
