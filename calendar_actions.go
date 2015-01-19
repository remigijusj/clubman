package main

import (
  "time"

  "github.com/gin-gonic/gin"
)

func redirectCalendar(c *gin.Context) {
  c.Redirect(302, "/calendar/week")
}

func getWeekData(c *gin.Context) {
  d := getDate(c)
  d = weekFirst(d) // monday

  c.Set("date", d)
  c.Set("prev", d.AddDate(0, 0, -7))
  c.Set("next", d.AddDate(0, 0, 7))
  c.Set("today", today())

  e := listWeekEventsGrouped(d)
  c.Set("events", e)
  t := indexTeams()
  c.Set("teams", t)
}

func getMonthData(c *gin.Context) {
  d := getDate(c)
  d = monthFirst(d)

  c.Set("date", d)
  c.Set("prev", d.AddDate(0, -1, 0))
  c.Set("next", d.AddDate(0, 1, 0))
  c.Set("today", today())

  e := listMonthEventsGrouped(d)
  c.Set("events", e)
  t := indexTeams()
  c.Set("teams", t)
}

// --- local helpers ---

func getDate(c *gin.Context) time.Time {
  date, err := time.Parse("2006-01-02", c.Request.URL.Query().Get("date"))
  if err != nil {
    date = today()
  }
  return date
}

func today() time.Time {
  return time.Now().UTC().Truncate(24 * time.Hour)
}
