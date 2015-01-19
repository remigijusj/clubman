package main

import (
  "fmt"

  "github.com/gin-gonic/gin"
)

func handleAssignmentCreate(c *gin.Context) {
  handleAssignmentAction(c, createAssignment, "Assignment has been created")
}

func handleAssignmentCancel(c *gin.Context) {
  handleAssignmentAction(c, cancelAssignment, "Assignment has been canceled")
}

func handleAssignmentDelete(c *gin.Context) {
  handleAssignmentAction(c, deleteAssignment, "Assignment has been deleted")
}

func getSelfAssignmentsList(c *gin.Context) {
  self := currentUser(c)
  if self == nil {
    forwardWarning(c, defaultPage, "Critical error happened, please contact website admin")
    return
  }
  list := listUserAssignments(self.Id)
  c.Set("list", list)
}

// --- local helpers ---

func handleAssignmentAction(c *gin.Context, action (func(int, int) error), message string) {
  event_id, err := getIntParam(c, "event_id")
  if err != nil {
    forwardWarning(c, defaultPage, err.Error())
    return
  }
  user_id, ok := getIntQuery(c, "user_id")
  if !ok {
    // action of self if no user_id
    if auth := currentUser(c); auth != nil {
      user_id = auth.Id
    }
  }
  err = action(event_id, user_id)
  if err != nil {
    forwardWarning(c, eventsViewPath(event_id), err.Error())
  } else {
    forwardTo(c, eventsViewPath(event_id), message)
  }
}

func eventsViewPath(event_id int) string {
  return fmt.Sprintf("/events/view/%d", event_id)
}
