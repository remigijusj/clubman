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
  prev := d.AddDate(0, 0, -7)
  next := d.AddDate(0, 0, 7)

  c.Set("date", d)
  c.Set("prev", prev)
  c.Set("next", next)
  c.Set("today", today())

  e := listWeekEventsGrouped(d)
  c.Set("events", e)
  t := indexTeams()
  c.Set("teams", t)

  if self := currentUser(c); self != nil {
    a := mapAssignedPeriod(self.Id, prev, next)
    c.Set("assigned", a)
  }
}

func getMonthData(c *gin.Context) {
  d := getDate(c)
  d = monthFirst(d)
  prev := d.AddDate(0, -1, 0)
  next := d.AddDate(0, 1, 0)

  c.Set("date", d)
  c.Set("prev", prev)
  c.Set("next", next)
  c.Set("today", today())

  e := listMonthEventsGrouped(d)
  c.Set("events", e)
  t := indexTeams()
  c.Set("teams", t)

  if self := currentUser(c); self != nil {
    a := mapAssignedPeriod(self.Id, prev, next)
    c.Set("assigned", a)
  }
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
