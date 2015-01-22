package main

import (
  "errors"
  "fmt"

  "github.com/gin-gonic/gin"
)

func handleAssignmentCreate(c *gin.Context) {
  handleAssignmentAction(c, createAssignment, "Assignment has been created")
}

func handleAssignmentDelete(c *gin.Context) {
  handleAssignmentAction(c, deleteAssignment, "Assignment has been deleted")
}

func getSelfAssignmentsList(c *gin.Context) {
  self := currentUser(c)
  if self == nil {
    forwardWarning(c, defaultPage, panicError)
    return
  }
  list := listUserAssignments(self.Id)
  c.Set("list", list)
}

func getUserAssignmentsList(c *gin.Context) {
  user_id, err := getIntParam(c, "id")
  if err != nil {
    forwardWarning(c, defaultPage, err.Error())
    return
  }
  list := listUserAssignments(user_id)
  c.Set("list", list)
  c.Set("id", user_id)
}

// --- local helpers ---

func handleAssignmentAction(c *gin.Context, action (func(int, int) error), message string) {
  event_id, err := getIntParam(c, "event_id")
  self := currentUser(c)
  if err != nil || self == nil {
    forwardWarning(c, defaultPage, err.Error())
    return
  }
  user_id, err := extractUserId(c, self)
  if err == nil {
    err = action(event_id, user_id)
  }
  if err != nil {
    forwardWarning(c, eventsViewPath(event_id), err.Error())
  } else {
    forwardTo(c, eventsViewPath(event_id), message)
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
