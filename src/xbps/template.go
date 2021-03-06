package xbps

import (
	"bytes"
	"exec"
	"fmt"
	"io"
	"json"
	"os"
	"path/filepath"
	"reflect"
)


type Template struct {
	Pkgname     string "pkgname"
	Version     string "version"
	Distfiles   string "distfiles"
	Short_Desc  string "short_desc"
	Maintainer  string "maintainer"
	Homepage    string "homepage"
	License     string "license"
	Checksum    string "checksum"
	Long_Desc   string "long_desc"
	Build_Style string "build_style"
}


func (this Template) ToSH() io.Reader {
	buf := new(bytes.Buffer)
	tmpl := reflect.ValueOf(this)
	nfield := tmpl.NumField()
	t := tmpl.Type()
	for i := 0; i < nfield; i++ {
		field := t.Field(i)
		value := tmpl.Field(i)
		if field.Tag != "" {
			fmt.Fprintf(buf, `%s="%s"%s`, field.Tag, value.String(), "\n")
		}
	}
	return buf
}

func (this Template) Save(file string) (err os.Error) {
	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	buf, err := this.ToJson()
	if err != nil {
		return
	}
	io.Copy(fd, buf)
	return
}

func (this Template) ToJson() (io.Reader, os.Error) {
	in := new(bytes.Buffer)
	err := json.NewEncoder(in).Encode(this)
	out := new(bytes.Buffer)
	err = json.Indent(out, in.Bytes(), "", "")
	return out, err
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
	dir, _ := filepath.Split(file)
	pname := filepath.Base(dir)
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

	template := new(Template)
	tmpl := reflect.ValueOf(*template)
	nfield := tmpl.NumField()
	t := tmpl.Type()
	buf.WriteString(`echo \{`)
	for i := 0; i <= nfield; i++ {
		field := t.Field(i)
		if field.Tag == "" {
			continue
		}
		line := fmt.Sprintf(`\"%s\":\"$%s\",`, field.Tag, field.Tag)
		if i == nfield-1 {
			line = line[0 : len(line)-1]
		}
		buf.WriteString(line)
	}

	buf.WriteString(`\}`)
	cmd := exec.Command("sh")
	cmd.Stdin = buf
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(output, template)
	if err != nil {
		return nil, fmt.Errorf("%-30.30s: %s", pname, err)
	}
	return template, nil
}

func GetTemplates(spath string) (map[string]*Template, os.Error) {
	os.Setenv("XBPS_SRCPKGDIR", spath)
	var templates = map[string]*Template{}
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
			continue
		}
		t, err := NewTemplate(file)
		if err != nil {
			return nil, err
		}
		if t != nil {
			templates[t.Pkgname] = t
		}
	}
	return templates, nil
}


func LoadJson(file string) (tmpl *Template, err os.Error) {
	tmpl = new(Template)
	fd, err := os.Open(file)
	if err != nil {
		return
	}
	defer fd.Close()
	err = json.NewDecoder(fd).Decode(tmpl)
	return
}
