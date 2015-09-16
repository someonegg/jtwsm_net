package present

import (
	"errors"
	"html/template"
	"path/filepath"
	"strings"
)

func init() {
	Register("html", parseHTML)
}

type HTML struct {
	template.HTML
}

func (s *HTML) TemplateName() string { return "html" }

func parseHTML(ctx *Context, fileName string, lineno int, text string) (Elem, error) {
	p := strings.Fields(text)
	if len(p) != 2 {
		return nil, errors.New("invalid .html args")
	}
	name := filepath.Join(filepath.Dir(fileName), p[1])
	b, err := ctx.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &HTML{template.HTML(b)}, nil
}

func init() {
	RegisterUnmarshaler("html", unmarshalHTML)
}

func unmarshalHTML(data []byte, worker UnmarshalWork) (Elem, error) {
	var ep struct {
		Inner *HTML
		Type  string
	}
	err := worker(data, &ep)
	if err != nil {
		return nil, err
	}
	return ep.Inner, nil
}
