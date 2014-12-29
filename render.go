package main

import (
  "html/template"
  "net/http"

  "github.com/gin-gonic/gin"
)

type (
  DevRender struct {
    Glob string
  }

  ProRender struct {
    Template *template.Template
  }
)

var helpers = template.FuncMap{
  "statusTitle": statusTitle,
  "statusList":  statusList,
}

// TODO: kill debug mode, or use gin.IsDebugging() later
const isDebugging = false

func loadHtmlTemplates(pattern string, engine *gin.Engine) {
  if isDebugging {
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
  obj := data[1]

  return r.Template.ExecuteTemplate(w, file, obj)
}

func (r DevRender) Render(w http.ResponseWriter, code int, data ...interface{}) error {
  writeHeader(w, code, "text/html")
  file := data[0].(string)
  obj := data[1]

  t := template.New("")
  if _, err := t.ParseGlob(r.Glob); err != nil {
    return err
  }
  return t.ExecuteTemplate(w, file, obj)
}

func writeHeader(w http.ResponseWriter, code int, contentType string) {
  w.Header().Set("Content-Type", contentType)
  w.WriteHeader(code)
}
