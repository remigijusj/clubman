package main

import (
  "fmt"

  "github.com/gin-gonic/gin"
)

type TransMap map[string]string

var translations = map[string]TransMap{}

func loadTranslations() {
  rows, err := query["translations"].Query()
  if err != nil { panic(err) }
  defer rows.Close()

  var locale string
  for rows.Next() {
    var key, value string
    err := rows.Scan(&locale, &key, &value)
    if err != nil { panic(err) }

    trans, exist := translations[locale]
    if !exist {
      trans = make(TransMap, 10)
      translations[locale] = trans
    }
    trans[key] = value
  }
  if err := rows.Err(); err != nil { panic(err) }
}

// TODO: interpolation
func T(lang, key string, args ...interface{}) string {
  if trans, ok := translations["da"]; ok {
    if val, ok := trans[key]; ok {
      return val
    }
  }
  if len(args) > 0 {
    key += fmt.Sprintf(" %v", args)// <<< DEBUG
  }
  return key
}

func TC(c *gin.Context, key string, args ...interface{}) string {
  return T(getLang(c), key, args...)
}
