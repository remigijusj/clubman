package main

import (
  "github.com/gin-gonic/gin"
)

func redirectCalendar(c *gin.Context) {
  c.Redirect(302, "/calendar/week")
}

func getWeekData(c *gin.Context) {
  date, _ := getDateQuery(c, "date")
  date = weekFirst(date) // monday
  c.Set("date", date)

  prev := date.AddDate(0, 0, -7)
  next := date.AddDate(0, 0, 7)
  c.Set("prev", prev)
  c.Set("next", next)

  team_id, _ := getIntQuery(c, "team_id")
  c.Set("team_id", team_id)

  events, eids := listWeekEventsGrouped(date, team_id)
  c.Set("events", events)

  setCommonCalendarData(c, eids)
}

func getMonthData(c *gin.Context) {
  date, _ := getDateQuery(c, "date")
  date = monthFirst(date)
  c.Set("date", date)

  prev := date.AddDate(0, -1, 0)
  next := date.AddDate(0, 1, 0)
  c.Set("prev", prev)
  c.Set("next", next)

  team_id, _ := getIntQuery(c, "team_id")
  c.Set("team_id", team_id)

  events, eids := listMonthEventsGrouped(date, team_id)
  c.Set("events", events)

  setCommonCalendarData(c, eids)
}

func setCommonCalendarData(c *gin.Context, eids []int) {
  c.Set("today", today())

  teams := indexTeams()
  c.Set("teams", teams)

  counts := mapParticipantCounts(eids)
  c.Set("counts", counts)

  if self := currentUser(c); self != nil {
    assigned := mapAssignedStatus(eids, self.Id)
    c.Set("assigned", assigned)
  }
}
