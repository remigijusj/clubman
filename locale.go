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

func T(lang, key string, args ...interface{}) string {
  if trans, ok := translations[lang]; ok {
    if val, ok := trans[key]; ok {
      return fmt.Sprintf(val, args...)
    }
  }
  return fmt.Sprintf(key, args...)
}

func TC(c *gin.Context, key string, args ...interface{}) string {
  return T(getLang(c), key, args...)
}
