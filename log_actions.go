package main

import (
  "github.com/gin-gonic/gin"
)

func getLogList(c *gin.Context) {
  q := c.Request.URL.Query()
  date, full := getDateQuery(c, "date")
  list := listLogsByQuery(q, date)
  c.Set("list", list)
  c.Set("full", full)
}
