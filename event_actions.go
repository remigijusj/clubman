package main

import (
  "errors"
  "fmt"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func handleEventsCreate(c *gin.Context) {
  handleEventsFormAction(c, createEvents, "%d events have been added")
}

func handleEventsCancel(c *gin.Context) {
  handleEventsFormAction(c, cancelEvents, "%d events have been canceled")
}

func handleEventsRemove(c *gin.Context) {
  handleEventsFormAction(c, removeEvents, "%d events have been removed")
}

func getEventForm(c *gin.Context) {
  var form EventForm
  event_id, err := getIntParam(c, "id")
  if err == nil {
    form, err = fetchEvent(event_id)
  }
  if err != nil {
    forwardWarning(c, defaultPage, err.Error())
    c.Abort(0)
  } else {
    setEventUpdateExtras(c, &form)
    c.Set("id", event_id)
    c.Set("form", form)
  }
}

func getEventAssignmentsList(c *gin.Context) {
  event_id, err := getIntParam(c, "id")
  if err != nil {
    forwardWarning(c, defaultPage, err.Error())
    c.Abort(0)
  } else {
    list := listEventAssignments(event_id)
    c.Set("list", list)
    user_names := mapUserNames(listAssignmentUserIds(list))
    c.Set("user_names", user_names)
    signed_up := findAssignment(list, currentUser(c))
    c.Set("signed_up", signed_up)
  }
}

func handleEventUpdate(c *gin.Context) {
  var form EventForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  event_id, err := getIntParam(c, "id")
  if err == nil {
    err = updateEvent(event_id, &form, getLang(c))
  }
  if err != nil {
    showError(c, err, &form)
  } else {
    forwardTo(c, eventsViewPath(event_id), "Event has been updated")
  }
}

func handleEventDelete(c *gin.Context) {
  event_id, err := getIntParam(c, "id")
  if err == nil {
    err = deleteEvent(event_id)
  }
  if err != nil {
    showError(c, err)
  } else {
    forwardTo(c, defaultPage, "Event has been deleted")
  }
}

// --- local helpers ---

func handleEventsFormAction(c *gin.Context, action (func(int, *TeamEventsForm, string) (int, error)), message string) {
  var cnt int
  team_id, err := getIntParam(c, "team_id")
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

func listAssignmentUserIds(list []AssignmentRecord) []int {
  ids := make([]int, len(list))
  for i, item := range list {
    ids[i] = item.UserId
  }
  return ids
}

// NOTE: this is needed for event update page only
func setEventUpdateExtras(c *gin.Context, form *EventForm) {
  lang := getLang(c)
  c.Set("date", time.Now().UTC().Format(dateFormats[lang]))

  form.Date = form.StartAt.Format(dateFormats[lang])
  form.Time = form.StartAt.Format(timeFormat)
}
