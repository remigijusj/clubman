package main

import (
  "time"

  "github.com/gin-gonic/gin"
)

func getWeekData(c *gin.Context) {
  d := getDate(c)
  d = d.Truncate(7 * 24 * time.Hour) // monday

  c.Set("date", d)
  c.Set("prev", d.AddDate(0, 0, -7))
  c.Set("next", d.AddDate(0, 0, 7))
}

func getMonthData(c *gin.Context) {
  d := getDate(c)
  d = d.AddDate(0, 0, -d.Day()) // 1st of month

  c.Set("date", d)
  c.Set("prev", d.AddDate(0, -1, 0))
  c.Set("next", d.AddDate(0, 1, 0))
}

func getDate(c *gin.Context) time.Time {
  date, err := time.Parse("2006-01-02", c.Request.URL.Query().Get("date"))
  if err != nil {
    date = time.Now()
  }
  return date
}
