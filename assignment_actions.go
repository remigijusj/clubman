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

// --- local helpers ---

func handleAssignmentAction(c *gin.Context, action (func(int, int) error), message string) {
  event_id, err := getIntParam(c, "event_id")
  if err != nil {
    forwardWarning(c, defaultPage, err.Error())
    return
  }
  path := fmt.Sprintf("/events/view/%d", event_id)
  user_id, ok := getIntQuery(c, "user_id")
  if !ok {
    // action of self if no user_id
    if auth := currentUser(c); auth != nil {
      user_id = auth.Id
    }
  }
  err = action(event_id, user_id)
  if err != nil {
    forwardWarning(c, path, err.Error())
  } else {
    forwardTo(c, path, message)
  }
}
