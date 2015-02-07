package main

import (
  "errors"
  "fmt"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func handleEventsCreate(c *gin.Context) {
  handleEventsFormAction(c, createEvents, "%d events have been created", "create")
}

func handleEventsUpdate(c *gin.Context) {
  handleEventsFormAction(c, updateEvents, "%d events have been updated", "update")
}

func handleEventsDelete(c *gin.Context) {
  handleEventsFormAction(c, deleteEvents, "%d events have been deleted", "delete")
}

func getEventForm(c *gin.Context) {
  var form EventForm
  event_id, err := getIntParam(c, "id")
  if err == nil {
    form, err = fetchEvent(event_id)
  }
  if err != nil {
    gotoWarning(c, defaultPage, err.Error())
    c.Abort(0)
  } else {
    setEventUpdateExtras(c, &form)
    c.Set("id", event_id)
    c.Set("form", form)
  }
}

func getEventTeam(c *gin.Context) {
  var team TeamForm
  form, err := c.Get("form")
  if event, ok := form.(EventForm); ok {
    team, err = fetchTeam(event.TeamId)
  }
  if err != nil {
    gotoWarning(c, defaultPage, panicError)
    c.Abort(0)
  } else {
    c.Set("team", team)
  }
}

func getEventAssignments(c *gin.Context) {
  event_id, err := getIntParam(c, "id")
  if err != nil {
    gotoWarning(c, defaultPage, err.Error())
    c.Abort(0)
  } else {
    list := listEventAssignments(event_id)
    c.Set("list", list)
    waiting := calcWaitingPosition(list)
    c.Set("waiting", waiting)
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
    gotoSuccess(c, eventsViewPath(event_id), "Event has been updated")
  }
}

func handleEventCancel(c *gin.Context) {
  event_id, err := getIntParam(c, "id")
  if err == nil {
    err = cancelEvent(event_id)
  }
  if err != nil {
    gotoWarning(c, eventsViewPath(event_id), err.Error())
  } else {
    gotoSuccess(c, eventsViewPath(event_id), "Event has been canceled")
  }
}

func handleEventDelete(c *gin.Context) {
  event_id, err := getIntParam(c, "id")
  if err == nil {
    err = deleteEvent(event_id)
  }
  if err != nil {
    gotoWarning(c, eventsViewPath(event_id), err.Error())
  } else {
    gotoSuccess(c, defaultPage, "Event has been deleted")
  }
}

func checkEventPerm(c *gin.Context) {
  if self := currentUser(c); self != nil {
    if self.IsAdmin() {
      return
    }
    if self.Status == userStatusInstructor {
      data, _ := c.Get("team")
      if team, ok := data.(TeamForm); ok {
        if team.InstructorId == self.Id {
          return
        }
      }
    }
  }
  gotoWarning(c, defaultPage, permitError)
  c.Abort(0)
}

func handleEventNotify(c *gin.Context) {
  subject := c.Request.FormValue("subject")
  message := c.Request.FormValue("message")
  if subject == "" || message == "" {
    showError(c, errors.New("Please provide all details"))
    return
  }
  var count int
  event_id, err := getIntParam(c, "id")
  if err == nil {
    count = notifyEventParticipants(event_id, subject, message)
  }
  if err != nil {
    showError(c, err)
  } else {
    gotoSuccess(c, eventsViewPath(event_id), "%d users have been notified", count)
  }
}

// --- local helpers ---

func handleEventsFormAction(c *gin.Context, action (func(int, *TeamEventsForm, string) (int, error)), message, tab string) {
  var cnt int
  team_id, err := getIntParam(c, "team_id")
  if err != nil {
    showError(c, err, nil, teamsEventsPath(team_id, tab))
    return
  }
  var form TeamEventsForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form, teamsEventsPath(team_id, tab))
    return
  }
  cnt, err = action(team_id, &form, getLang(c))
  if err != nil {
    showError(c, err, &form, teamsEventsPath(team_id, tab))
  } else {
    gotoSuccess(c, teamsEventsPath(team_id, tab), message, cnt)
  }
}

func teamsEventsPath(team_id int, tab string) string {
  return fmt.Sprintf("/teams/events/%d#%s", team_id, tab)
}

// NOTE: this is needed for event update page only
func setEventUpdateExtras(c *gin.Context, form *EventForm) {
  lang := getLang(c)
  c.Set("date", today().Format(dateFormats[lang]))

  form.Date = form.StartAt.Format(dateFormats[lang])
  form.Time = form.StartAt.Format(timeFormat)
}
