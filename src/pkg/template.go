package xbps

import (
	"bytes"
	"exec"
	"fmt"
	"io"
	"json"
	"os"
	"path/filepath"
)

var shVars = []string{
	"pkgname",
	"version",
	"distfiles",
	"maintainer",
	"homepage",
	"license",
	"checksum",
}

type Template struct {
	Pkgname    string
	Version    string
	Distfiles  string
	ShortDesc  string "short_desc"
	Maintainer string
	Homepage   string
	License    string
	Checksum   string
	LongDesc   *json.RawMessage "long_desc"
	Path       string
}

func FindTemplate(pkg, spath string) (tmpl *Template, err os.Error) {
	tpath := fmt.Sprintf("%s/%s/template", spath, pkg)
	_, err = os.Stat(tpath)
	if err != nil {
		return
	}
	tmpl, err = NewTemplate(tpath)
	if err != nil {
		return
	}
	return
}

func NewTemplate(file string) (*Template, os.Error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	fs, err := os.Open("/usr/local/share/xbps-src/common/fetch_sites.sh")
	if err != nil {
		return nil, err
	}
	defer fs.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, fs)
	buf.WriteString("Add_dependency(){ :\n }\n")
	io.Copy(buf, fd)
	buf.WriteString("echo {")
	for i, v := range shVars {
		line := fmt.Sprintf(`\"%s\":\"$%s\",`, v, v)
		if i == len(shVars)-1 {
			line = line[0 : len(line)-1]
		}
		buf.WriteString(line)
	}
	buf.WriteString("}")
	cmd := exec.Command("sh")
	cmd.Stdin = buf
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	template := new(Template)
	err = json.Unmarshal(output, template)
	if err != nil {
		return nil, err
	}
	template.Path = file
	return template, nil
}

func GetTemplates(spath string) (map[string]*Template, os.Error) {
	os.Setenv("XBPS_SRCPKGDIR", spath)
	var (
		linked    = 0
		ok        = 0
		templates = map[string]*Template{}
	)
	files, err := filepath.Glob(spath + "/*/template")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		d, _ := filepath.Split(file)
		d = filepath.Clean(d)
		di, err := os.Stat(d)
		if err != nil {
			return nil, err
		}
		if di.FollowedSymlink {
			linked++
			continue
		}
		t, err := NewTemplate(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ok++
		if t != nil {
			templates[t.Pkgname] = t
		}
	}
	/*
		total := errors + linked + ok
		fmt.Println("errors",errors)
		fmt.Println("linked",linked)
		fmt.Println("ok",ok)
		fmt.Println("total",total)
	*/
	return templates, nil
}
