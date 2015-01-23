package main

import (
  "github.com/gin-gonic/gin"
)

func redirectCalendar(c *gin.Context) {
  c.Redirect(302, "/calendar/week")
}

func getWeekData(c *gin.Context) {
  d, _ := getDateQuery(c, "date")
  d = weekFirst(d) // monday
  prev := d.AddDate(0, 0, -7)
  next := d.AddDate(0, 0, 7)

  c.Set("date", d)
  c.Set("prev", prev)
  c.Set("next", next)
  c.Set("today", today())

  team_id, _ := getIntQuery(c, "team_id")
  c.Set("team_id", team_id)

  e := listWeekEventsGrouped(d, team_id)
  c.Set("events", e)
  t := indexTeams()
  c.Set("teams", t)

  if self := currentUser(c); self != nil {
    a := mapAssignedPeriod(self.Id, prev, next)
    c.Set("assigned", a)
  }
}

func getMonthData(c *gin.Context) {
  d, _ := getDateQuery(c, "date")
  d = monthFirst(d)
  prev := d.AddDate(0, -1, 0)
  next := d.AddDate(0, 1, 0)

  team_id, _ := getIntQuery(c, "team_id")
  c.Set("team_id", team_id)

  c.Set("date", d)
  c.Set("prev", prev)
  c.Set("next", next)
  c.Set("today", today())

  e := listMonthEventsGrouped(d, team_id)
  c.Set("events", e)
  t := indexTeams()
  c.Set("teams", t)

  if self := currentUser(c); self != nil {
    a := mapAssignedPeriod(self.Id, prev, next)
    c.Set("assigned", a)
  }
}
