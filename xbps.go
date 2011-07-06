package xbps

import (
	"bytes"
	"exec"
	"fmt"
	. "github.com/str1ngs/go-ansi/color"
	"io"
	"log"
	"os"
	"strings"
)

var (
	xbpsBin  = "xbps-bin.static -c %s -Ayr %s %s %s"
	xbpsSrc  = "xbps-src -C -m %s %s %s"
	binCache = "/home/strings/bincache"
	debug    = false
	bufout   = new(bytes.Buffer)
)


func HandleError(err os.Error) {
	log.Print(err)
	log.Print("*****************************************************************")
	io.Copy(os.Stderr, bufout)
	log.Print("*****************************************************************")
}

func Seed(md *MasterDir) (err os.Error) {
	err = Install("base-chroot>=0.11", md)
	if err != nil {
		md.UnMount()
		return err
	}
	return
}

func Install(pack string, md *MasterDir) (err os.Error) {
	if IsInstalled(pack, md) {
		log.Printf("%-10.10s %s", "skip", pack)
		return
	}
	log.Printf("%-10.10s %s", "install", pack)
	err = NewCommand(xbpsBin, binCache, md.TargetPath, "install", pack).Run()
	if err != nil {
		return
	}
	return
}

func Build(tmpl string, md *MasterDir) (err os.Error) {
	depends, err := GetDepends(tmpl)
	if err != nil {
		return err
	}
	for _, d := range depends {
		err = Install(d, md)
		if err != nil {
			return err
		}
	}
	log.Printf("%-10.10s %s", "build", tmpl)
	cmd := NewCommand(xbpsSrc, md.TargetPath, "install", tmpl)
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

func Package(tmpl string, md *MasterDir) (err os.Error) {
	log.Printf("%-10.10s %s", "package", tmpl)
	err = NewCommand(xbpsSrc, md.TargetPath, "build-pkg", tmpl).Run()
	if err != nil {
		return
	}

	log.Printf("%-10.10s %s", "index", tmpl)
	err = NewCommand(xbpsSrc, md.TargetPath, "make-repoidx", tmpl).Run()
	if err != nil {
		return
	}

	log.Printf("%-10.10s %s", "remove", tmpl)
	err = NewCommand(xbpsSrc, md.TargetPath, "remove", tmpl).Run()
	if err != nil {
		return
	}

	if err := md.UnMount(); err != nil {
		return err
	}

	log.Printf("%-10.10s %s", "removing", md.TargetPath)
	err = os.RemoveAll(md.TargetPath)
	if err != nil {
		return err
	}
	log.Printf("%-10.10s %s", "done", tmpl)
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
