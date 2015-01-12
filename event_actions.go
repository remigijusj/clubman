package main

import (
  "errors"
  "fmt"
  //"log"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func handleEventsAdd(c *gin.Context) {
  handleEventsFormAction(c, addEvents, "%d events have been added")
}

func handleEventsCancel(c *gin.Context) {
  handleEventsFormAction(c, cancelEvents, "%d events have been canceled")
}

func handleEventsRemove(c *gin.Context) {
  handleEventsFormAction(c, removeEvents, "%d events have been removed")
}

// --- local helpers ---

func handleEventsFormAction(c *gin.Context, action (func(int, *TeamEventsForm, string) (int, error)), message string) {
  var cnt int
  team_id, err := teamId(c)
  if err != nil {
    showError(c, err, nil, teamsEventsPath(team_id))
    return
  }
  var form TeamEventsForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form, teamsEventsPath(team_id))
    return
  }
  cnt, err = action(team_id, &form, getLang(c))
  if err != nil {
    showError(c, err, &form, teamsEventsPath(team_id))
  } else {
    forwardTo(c, teamsEventsPath(team_id), message, cnt)
  }
}

func teamsEventsPath(team_id int) string {
  return fmt.Sprintf("/teams/events/%d", team_id)
}
