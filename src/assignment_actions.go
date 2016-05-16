package main

import (
  "database/sql"
  "errors"
  "fmt"

  "github.com/gin-gonic/gin"
)

func getSelfAssignmentsList(c *gin.Context) {
  self := currentUser(c)
  if self == nil {
    gotoWarning(c, conf.DefaultPage, panicError)
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
    gotoWarning(c, conf.DefaultPage, err.Error())
    return
  }
  date, full := getDateQuery(c, "date")
  list := listUserAssignments(user_id, date)
  c.Set("list", list)
  c.Set("full", full)
  c.Set("id", user_id)
}

func handleAssignmentCreate(c *gin.Context) {
  event_id, user_id, tx := prepareAssignmentAction(c)
  if tx == nil { return }

  status, err := decideAssignmentStatus(event_id, tx)
  if err != nil {
    failAssignmentAction(c, event_id, tx, "")
    return
  }
  err = createAssignmentTx(tx, event_id, user_id, status)
  if err != nil {
    failAssignmentAction(c, event_id, tx, "Subscription failed, perhaps user is already subscribed")
    return
  }

  message := assignmentCreateSuccess(c, user_id, status)
  if ok := completeAssignmentAction(c, event_id, tx, message); ok {
    go func() {
      notifyAssignmentAction(event_id, user_id, status)
    }()
  }
}

func handleAssignmentDelete(c *gin.Context) {
  event_id, user_id, tx := prepareAssignmentAction(c)
  if tx == nil { return }

  err := deleteAssignmentTx(tx, event_id, user_id)
  if err != nil {
    failAssignmentAction(c, event_id, tx, "")
    return
  }

  message := assignmentDeleteSuccess(c, user_id)
  if ok := completeAssignmentAction(c, event_id, tx, message); ok {
    go func() {
      notifyAssignmentAction(event_id, user_id, 0)
      afterAssignmentDelete(event_id, 0)
    }()
  }
}

func handleAssignmentConfirm(c *gin.Context) {
  event_id, user_id, tx := prepareAssignmentAction(c)
  if tx == nil { return }

  err := confirmAssignmentTx(tx, event_id, user_id)
  if err != nil {
    failAssignmentAction(c, event_id, tx, "Action failed, perhaps the subscription is already confirmed or canceled")
    return
  }

  message := assignmentConfirmSuccess(c, user_id)
  completeAssignmentAction(c, event_id, tx, message)
}

// --- local helpers ---

func prepareAssignmentAction(c *gin.Context) (int, int, *sql.Tx) {
  event_id, err := getIntParam(c, "event_id")
  self := currentUser(c)
  if err != nil || self == nil {
    gotoWarning(c, conf.DefaultPage, err.Error())
    return 0, 0, nil
  }
  user_id, err := extractUserId(c, self)
  if err != nil {
    gotoWarning(c, eventsViewPath(event_id), err.Error())
    return 0, 0, nil
  }

  tx, err := db.Begin()
  if err != nil {
    failAssignmentAction(c, event_id, nil, "")
    return 0, 0, nil
  }
  return event_id, user_id, tx
}

func completeAssignmentAction(c *gin.Context, event_id int, tx *sql.Tx, message string) bool {
  err := tx.Commit()
  if err != nil {
    failAssignmentAction(c, event_id, nil, "")
    return false
  }
  gotoSuccess(c, eventsViewPath(event_id), message)
  return true
}

func failAssignmentAction(c *gin.Context, event_id int, tx *sql.Tx, message string) {
  if tx != nil {
    tx.Rollback()
  }
  if message == "" {
    message = "Action could not be completed, please try again later"
  }
  gotoWarning(c, eventsViewPath(event_id), message)
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

func decideAssignmentStatus(event_id int, tx *sql.Tx) (int, error) {
  max, err := maxTeamUsersTx(tx, event_id)
  if err != nil {
    return 0, err
  }
  if max == 0 {
    return assignmentStatusConfirmed, nil
  }
  cnt, err := countAssignmentsTx(tx, event_id)
  if err != nil {
    return 0, err
  }
  if cnt >= max {
    return assignmentStatusWaiting, nil
  }
  return assignmentStatusConfirmed, nil
}

func assignmentCreateSuccess(c *gin.Context, user_id, status int) string {
  self := currentUser(c)
  if self == nil {
    return panicError
  }
  switch status {
  case assignmentStatusConfirmed:
    if self.Id == user_id {
      return "Your subscription has been confirmed"
    } else {
      return "User has been subscribed"
    }
  case assignmentStatusWaiting:
    if self.Id == user_id {
      return "You have been put on the waiting list. If somebody unsubscribes you will be notified"
    } else {
      return "User has been put on the waiting list"
    }
  default:
    return panicError
  }
}

func assignmentDeleteSuccess(c *gin.Context, user_id int) string {
  self := currentUser(c)
  if self == nil {
    return panicError
  }
  if self.Id == user_id {
    return "Your subscription has been canceled"
  } else {
    return "User subscription has been canceled"
  }
}

func assignmentConfirmSuccess(c *gin.Context, user_id int) string {
  self := currentUser(c)
  if self == nil {
    return panicError
  }
  if self.Id == user_id {
    return "Your subscription has been confirmed"
  } else {
    return "User subscription has been confirmed"
  }
}
