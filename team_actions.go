package main

import (
  "errors"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func getTeamList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listTeamsByQuery(q)
  c.Set("list", list)
}

func getTeamEventsData(c *gin.Context) {
  team_id, err := getIntParam(c, "id")
  if err != nil {
    forwardWarning(c, "/teams", err.Error())
    c.Abort(0)
  } else {
    list := listTeamEvents(team_id)
    c.Set("list", list)
    counts := countEventAssignments(list)
    c.Set("counts", counts)
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
    forwardWarning(c, "/teams", err.Error())
    c.Abort(0)
  } else {
    c.Set("id", team_id)
    c.Set("form", form)
  }
}

func handleTeamCreate(c *gin.Context) {
  var form TeamForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  err := createTeam(&form)
  if err != nil {
    showError(c, err, &form)
  } else {
    forwardTo(c, "/teams", "Team has been created")
  }
}

func handleTeamUpdate(c *gin.Context) {
  var form TeamForm
  if ok := c.BindWith(&form, binding.Form); !ok {
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
    forwardTo(c, "/teams", "Team has been updated")
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
    forwardTo(c, "/teams", "Team has been deleted")
  }
}

func newTeamEventsForm(c *gin.Context) {
  team, _ := c.Get("form")
  c.Set("team", team)
  form := TeamEventsForm{}
  c.Set("form", form)
  // placeholders
  lang := getLang(c)
  c.Set("date_from", time.Now().UTC().Format(dateFormats[lang]))
  c.Set("date_till", time.Date(time.Now().UTC().Year()+1, 1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormats[lang]))
}
