package handlers

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"

	"goapp/internal/i18n"
)

type TemplateRenderer struct {
	templates *template.Template
}

func NewTemplateRenderer(pattern string) *TemplateRenderer {
	funcs := template.FuncMap{"t": i18n.T}
	return &TemplateRenderer{templates: template.Must(template.New("").Funcs(funcs).ParseGlob(pattern))}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
