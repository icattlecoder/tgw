package tgw

import (
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
)

type view struct {
	viewDir string
	delims  []string
	cache   map[string]*template.Template
}

func NewView(viewDir string) (v *view, err error) {
	f, err := os.Stat(viewDir)
	if err != nil {
		return
	}
	if !f.IsDir() {
		err = errors.New(viewDir + " is not a Dir")
	}
	cache := map[string]*template.Template{}
	delims := []string{"((", "))"}
	v = &view{viewDir: viewDir, cache: cache, delims: delims}
	return
}

// Example usage: <include src="include/header.hmtl" />
func (v *view) includeHandler(content string) (result string) {
	result = content
	r := regexp.MustCompile(`<include src="([^>]*)" />`)
	matches := r.FindAllStringSubmatch(result, -1)
	for _, val := range matches {
		sr := regexp.MustCompile(val[0])
		icld, err := v.readViewFile(val[1])
		if err == nil {
			result = sr.ReplaceAllString(result, icld)
		}
	}
	return
}

func (v *view) readViewFile(filename string) (content string, err error) {
	pat := path.Join(v.viewDir, filename)
	bs, err := ioutil.ReadFile(pat)
	if err != nil {
		return
	}
	content = string(bs)
	return
}

func (v *view) GetHtml(name string) (reader io.Reader, err error) {
	//html的后缀有个好处，sublime等编辑器可对其代码格式化
	name += ".html"
	pat := path.Join(v.viewDir, name)
	reader, err = os.Open(pat)
	return
}

func (v *view) Get(name string) (tpl *template.Template, err error) {
	//html的后缀有个好处，sublime等编辑器可对其代码格式化
	name += ".html"
	pat := path.Join(v.viewDir, name)
	if DEBUG {
		icld := ""
		icld, err = v.readViewFile(name)
		if err != nil {
			return nil, err
		}
		icld = v.includeHandler(icld)
		tpl, err = template.New(name).Delims(v.delims[0], v.delims[1]).Parse(icld)
		if err != nil {
			log.Println("Template.Parse err:", err)
		}
		return
	}
	if tpl2, ok := v.cache[pat]; ok {
		tpl = tpl2
		return
	} else {
		icld := ""
		icld, err = v.readViewFile(name)
		if err != nil {
			return nil, err
		}
		icld = v.includeHandler(icld)
		tpl, err = template.New(name).Delims(v.delims[0], v.delims[1]).Parse(icld)
		if err != nil {
			return
		}
		v.cache[pat] = tpl
		return
	}
}
