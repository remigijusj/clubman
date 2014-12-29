package main

import (
  "html/template"
  "log"
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

func loadTemplates(engine *gin.Engine, pattern string) {
  if true { // <<< gin.IsDebugging()
    engine.HTMLRender = DevRender{
      Glob: pattern,
    }
  } else {
    engine.HTMLRender = ProRender{
      Template: template.Must(template.ParseGlob(pattern)),
    }
  }
}

func (html ProRender) Render(w http.ResponseWriter, code int, data ...interface{}) error {
  writeHeader(w, code, "text/html")
  file := data[0].(string)
  obj := data[1]

  return html.Template.ExecuteTemplate(w, file, obj)
}

func (r DevRender) Render(w http.ResponseWriter, code int, data ...interface{}) error {
  writeHeader(w, code, "text/html")
  file := data[0].(string)
  obj := data[1]

  t := template.New("")
  if _, err := t.ParseGlob(r.Glob); err != nil {
    return err
  }
for _, x := range t.Templates() {
  log.Printf("=> T: %#v\n", x.Name())
}
  return t.ExecuteTemplate(w, file, obj)
}

func writeHeader(w http.ResponseWriter, code int, contentType string) {
  w.Header().Set("Content-Type", contentType)
  w.WriteHeader(code)
}
