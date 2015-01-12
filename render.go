package main

import (
  "html/template"
  "net/http"

  "github.com/gin-gonic/gin"
)

type DevRender struct {
  Glob string
}

type ProRender struct {
  Template *template.Template
}

var helpers = template.FuncMap{
  "containsInt": containsInt,
  "listRecords": listRecords,
  "statusTitle": statusTitle,
  "statusList":  statusList,
  "userName":    userName,
  "serverName": func() string { return serverName },
  "timeFormat": func() string { return timeFormat },
  "dateFormat": func(lang string) string { return dateFormats[lang] },
  "T": func(key string) string { return key },
}

func loadHtmlTemplates(pattern string, engine *gin.Engine) {
  if reloadTmpl {
    engine.HTMLRender = DevRender{
      Glob: pattern,
    }
  } else {
    tmpl, _ := template.New("").Funcs(helpers).ParseGlob(pattern)
    engine.HTMLRender = ProRender{
      Template: tmpl,
    }
  }
}

func (r ProRender) Render(w http.ResponseWriter, code int, data ...interface{}) error {
  writeHeader(w, code, "text/html")
  file := data[0].(string)
  obj := data[1].(gin.H)

  addTranslations(r.Template, obj)

  return r.Template.ExecuteTemplate(w, file, obj)
}

func (r DevRender) Render(w http.ResponseWriter, code int, data ...interface{}) error {
  writeHeader(w, code, "text/html")
  file := data[0].(string)
  obj := data[1].(gin.H)

  t := template.New("").Funcs(helpers)
  if _, err := t.ParseGlob(r.Glob); err != nil {
    return err
  }
  addTranslations(t, obj)

  return t.ExecuteTemplate(w, file, obj)
}

func writeHeader(w http.ResponseWriter, code int, contentType string) {
  w.Header().Set("Content-Type", contentType)
  w.WriteHeader(code)
}

func addTranslations(t *template.Template, obj gin.H) {
  trans := transHelpers[defaultLang]
  if lang, ok := obj["lang"].(string); ok {
    trans = transHelpers[lang]
  }
  t.Funcs(trans)
}
