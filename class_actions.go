package main

import (
  "github.com/gin-gonic/gin"
)

func getClassesList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listClasses(q)
  c.Set("list", list)
}
