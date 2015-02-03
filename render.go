package main

import (
  "html/template"
  "net/http"
  "time"

  "github.com/gin-gonic/gin"
)

type DevRender struct {
  Glob string
}

type ProRender struct {
  Template *template.Template
}

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
  "defaultDate": func() string { return defaultDate },
  "printTime":   func(t time.Time) string { return t.Format(timeFormat) },
  "printDate":   func(t time.Time) string { return t.Format(dateFormat) },
  "localDate":   func(t time.Time, lang string) string { return t.Format(dateFormats[lang]) },
  "truncate":    func(i int, s string) string { return s[0:i] },
  "T":           func(key string) string { return key },
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
