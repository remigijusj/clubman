package main

import (
  "fmt"
  "html/template"

  "github.com/gin-gonic/gin"
)

var translations = map[string]map[string]string{}

var transHelpers = map[string]template.FuncMap{}

func loadTranslations() {
  for lang, _ := range locales {
    translations[lang] = make(map[string]string)
  }

  rows, err := query["translations"].Query()
  if err != nil { panic(err) }
  defer rows.Close()

  var lang string
  for rows.Next() {
    var key, value string
    err := rows.Scan(&lang, &key, &value)
    if err != nil { panic(err) }

    trans, exist := translations[lang]
    if !exist || value == "" { continue }
    trans[key] = value
  }
  if err := rows.Err(); err != nil { panic(err) }
}

func makeTransHelpers() {
  for lang, _ := range locales {
    trans := translations[lang]
    helper := func(key string, args ...interface{}) string {
      if val, ok := trans[key]; ok { key = val }
      return fmt.Sprintf(key, args...)
    }
    transHelpers[lang] = template.FuncMap{
      "T": helper,
    }
  }
}

func T(lang, key string, args ...interface{}) string {
  if trans, ok := translations[lang]; ok {
    if val, ok := trans[key]; ok {
      key = val
    }
  }
  return fmt.Sprintf(key, args...)
}

func TC(c *gin.Context, key string, args ...interface{}) string {
  return T(getLang(c), key, args...)
}