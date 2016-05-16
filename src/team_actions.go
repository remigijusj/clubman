package main

import (
  "errors"

  "github.com/gin-gonic/gin"
)

func getTeamList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listTeamsByQuery(q)
  c.Set("list", list)
}

func getTeamEventsData(c *gin.Context) {
  team_id, err := getIntParam(c, "id")
  if err != nil {
    gotoWarning(c, "/teams", err.Error())
    c.Abort()
    return
  }
  date, full := getDateQuery(c, "date")
  list := listTeamEvents(team_id, date)
  eids := collectEventIds(list)

  c.Set("list", list)
  c.Set("full", full)

  counts := mapParticipantCounts(eids)
  c.Set("counts", counts)

  if self := currentUser(c); self != nil {
    assigned := mapAssignedStatus(eids, self.Id)
    c.Set("assigned", assigned)
  }
}

func newTeamForm(c *gin.Context) {
  form := TeamForm{}
  c.Set("form", form)
}

func getTeamForm(c *gin.Context) {
  var form TeamForm
  team_id, err := getIntParam(c, "id")
  if err == nil {
    form, err = fetchTeam(team_id)
  }
  if err != nil {
    gotoWarning(c, "/teams", err.Error())
    c.Abort()
  } else {
    c.Set("id", team_id)
    c.Set("form", form)
  }
}

func handleTeamCreate(c *gin.Context) {
  var form TeamForm
  if ok := bindForm(c, &form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  err := createTeam(&form)
  if err != nil {
    showError(c, err, &form)
  } else {
    gotoSuccess(c, "/teams", "Team has been created")
  }
}

func handleTeamUpdate(c *gin.Context) {
  var form TeamForm
  if ok := bindForm(c, &form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  team_id, err := getIntParam(c, "id")
  if err == nil {
    err = updateTeam(team_id, &form)
  }
  if err != nil {
    showError(c, err, &form)
  } else {
    gotoSuccess(c, "/teams", "Team has been updated")
  }
}

func handleTeamDelete(c *gin.Context) {
  team_id, err := getIntParam(c, "id")
  if err == nil {
    err = deleteTeam(team_id)
  }
  if err != nil {
    showError(c, err)
  } else {
    gotoSuccess(c, "/teams", "Team has been deleted")
  }
}

func newTeamEventsForm(c *gin.Context) {
  team, _ := c.Get("form")
  c.Set("team", team)
  form := TeamEventsForm{}
  c.Set("form", form)
  // placeholders
  lang := getLang(c)
  c.Set("date_from", today().Format(locales[lang].Date))
  // WAS: time.Date(today().Year()+1, 1, 0, 0, 0, 0, 0, time.UTC).Format(locales[lang].Date)
  c.Set("date_till", "")
}
