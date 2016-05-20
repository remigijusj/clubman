package main

import (
  "github.com/gin-gonic/gin"
)

func getTranslationList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listTranslationsByQuery(q)
  c.Set("list", list)
}
