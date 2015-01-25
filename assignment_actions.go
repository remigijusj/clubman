package main

import (
  "errors"
  "fmt"

  "github.com/gin-gonic/gin"
)

func getSelfAssignmentsList(c *gin.Context) {
  self := currentUser(c)
  if self == nil {
    gotoWarning(c, defaultPage, panicError)
    return
  }
  date, full := getDateQuery(c, "date")
  list := listUserAssignments(self.Id, date)
  c.Set("list", list)
  c.Set("full", full)
}

func getUserAssignmentsList(c *gin.Context) {
  user_id, err := getIntParam(c, "id")
  if err != nil {
    gotoWarning(c, defaultPage, err.Error())
    return
  }
  date, full := getDateQuery(c, "date")
  list := listUserAssignments(user_id, date)
  c.Set("list", list)
  c.Set("full", full)
  c.Set("id", user_id)
}

func handleAssignmentCreate(c *gin.Context) {
  handleAssignmentAction(c, createAssignment)
}

func handleAssignmentDelete(c *gin.Context) {
  handleAssignmentAction(c, deleteAssignment)
}

// --- local helpers ---

func handleAssignmentAction(c *gin.Context, action (func(int, int) (bool, string))) {
  event_id, err := getIntParam(c, "event_id")
  self := currentUser(c)
  if err != nil || self == nil {
    gotoWarning(c, defaultPage, err.Error())
    return
  }
  user_id, err := extractUserId(c, self)
  if err != nil {
    gotoWarning(c, eventsViewPath(event_id), err.Error())
    return
  }
  ok, message := action(event_id, user_id)
  if ok {
    gotoSuccess(c, eventsViewPath(event_id), message)
  } else {
    gotoWarning(c, eventsViewPath(event_id), message)
  }
}

func extractUserId(c *gin.Context, self *AuthInfo) (int, error) {
  if isAdmin(c) {
    if user_id, ok := getIntQuery(c, "user_id"); ok {
      if user_id > 0 {
        return user_id, nil
      } else {
        return 0, errors.New("Please choose a user to signup")
      }
    }
  }
  return self.Id, nil
}

func eventsViewPath(event_id int) string {
  return fmt.Sprintf("/events/view/%d", event_id)
}
