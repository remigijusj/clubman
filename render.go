package main

import (
  "html/template"
  "net/http"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/render"
)

type PageTemplate struct {
  TemplateGlob string
  templates    *template.Template
}

type PageRender struct {
  Template *template.Template
  Data     interface{}
  Name     string
}

var htmlContentType = []string{"text/html; charset=utf-8"}

var helpers = template.FuncMap{
  "calcMonthDate": calcMonthDate,
  "containsInt":   containsInt,
  "listRecords":   listRecords,
  "statusTitle":   statusTitle,
  "statusList":    statusList,
  "eventClass":    eventClass,
  "userName":      userName,
  "dict":          dict,
  "serverName":  func() string { return serverName },
  "adminEmail":  func() string { return adminEmail },
  "defaultDate": func() string { return defaultDate },
  "printTime":   func(t time.Time) string { return t.Format(timeFormat) },
  "printDate":   func(t time.Time) string { return t.Format(dateFormat) },
  "localDate":   func(t time.Time, lang string) string { return t.Format(dateFormats[lang]) },
  "truncate":    func(i int, s string) string { return s[:i] },
  "taketill":    func(delim, s string) string { i := strings.Index(s, delim); if i < 0 { i = len(s) }; return s[:i] },
  "T":           func(key string) string { return key },
}

func loadHtmlTemplates(pattern string, engine *gin.Engine) {
  tmpl := PageTemplate{pattern, nil}
  tmpl.loadTemplates()
  engine.HTMLRender = tmpl
}

func (r *PageTemplate) loadTemplates() error {
  tmpl, err := template.New("").Funcs(helpers).ParseGlob(r.TemplateGlob)
  r.templates = tmpl
  return err
}

func (r PageTemplate) Instance(name string, data interface{}) render.Render {
  if debugMode {
    r.loadTemplates()
  }
  return PageRender{
    Template: r.templates,
    Name:     name,
    Data:     data,
  }
}

func (r PageRender) Render(w http.ResponseWriter) error {
  writeContentType(w, htmlContentType)

  setTranslations(r.Template, r.Data)

  if len(r.Name) > 0 {
    r.Template.ExecuteTemplate(w, r.Name, r.Data)
  } else {
    r.Template.Execute(w, r.Data)
  }

  return nil
}

func setTranslations(t *template.Template, data interface{}) {
  trans := transHelpers[defaultLang]
  if obj, ok := data.(gin.H); ok {
    if lang, ok := obj["lang"].(string); ok {
      trans = transHelpers[lang]
    }
  }
  t.Funcs(trans)
}

func writeContentType(w http.ResponseWriter, value []string) {
  header := w.Header()
  if val := header["Content-Type"]; len(val) == 0 {
    header["Content-Type"] = value
  }
}
