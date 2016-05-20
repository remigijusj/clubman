package main

import (
  "github.com/gin-gonic/gin"
)

func getTranslationList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listTranslationsByQuery(q)
  c.Set("list", list)
}

func getTranslationForm(c *gin.Context) {
  var form TranslationForm
  rowid, err := getIntParam(c, "id")
  if err == nil {
    form, err = fetchTranslation(rowid)
  }
  if err != nil {
    gotoWarning(c, "/translations", err.Error())
    c.Abort()
  } else {
    c.Set("rowid", rowid)
    c.Set("form",  form)
  }
}
