package main

import (
  "database/sql"
  "errors"
  "fmt"
  "log"
)

// NOTE: delayed
func afterAssignmentDelete(event_id int, lang string) {
  tx, err := db.Begin()
  if err != nil { return }

  user, err := firstWaitingUserTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  err = updateAssignmentStatusTx(tx, event_id, user.Id, assignmentStatusNotified)
  if err != nil { tx.Rollback(); return }

  err = tx.Commit()
  if err != nil { return }

  sendWaitingConfirmationEmail(lang, user.Email, user.Id, event_id)
}

func firstWaitingUserTx(tx *sql.Tx, event_id int) (*UserContact, error) {
  max, err := maxTeamUsersTx(tx, event_id)
  if err != nil {
    return nil, err
  }

  cnt, err := countNonWaitingUsersTx(tx, event_id)
  if err != nil {
    return nil, err
  }
  if cnt >= max {
    log.Printf("[APP] ASSIGNMENTS-FIRST-WAITING condition: %d\n", event_id)
    return nil, errors.New("")
  }

  var user UserContact
  err = tx.Stmt(query["assignments_first"]).QueryRow(event_id, assignmentStatusWaiting).Scan(&user.Id, &user.Email, &user.Mobile)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-FIRST-WAITING error: %s, %d, %v\n", err, event_id, user)
  }
  return &user, err
}

// NOTE: count status IN (-1, 1)
func countNonWaitingUsersTx(tx *sql.Tx, event_id int) (int, error) {
  var count int
  err := tx.Stmt(query["assignments_check"]).QueryRow(event_id, assignmentStatusConfirmed, assignmentStatusNotified).Scan(&count)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-CHECK-STATUS error: %s, %d\n", err, event_id)
  }
  return count, err
}

func updateAssignmentStatusTx(tx *sql.Tx, event_id, user_id, status int) error {
  res, err := tx.Stmt(query["assignment_update"]).Exec(status, event_id, user_id)
  if err != nil {
    log.Printf("[APP] ASSIGNMENT-UPDATE-STATUS error: %s, %d, %d\n", err, event_id, user_id)
    return err
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return err
  }
  return nil
}

func confirmAssignmentTx(tx *sql.Tx, event_id, user_id int) (err error) {
  defer func() {
    log.Printf("[APP] ASSIGNMENT-CONFIRM-STATUS error: %s, %d, %d\n", err, event_id, user_id)
  }()

  var status int
  err = tx.Stmt(query["assignments_status"]).QueryRow(event_id, user_id).Scan(&event_id, &status)
  if err != nil { return }

  if status != assignmentStatusNotified {
    err = errors.New(fmt.Sprintf("invalid status: %d", status))
  }
  if err != nil { return }

  err = updateAssignmentStatusTx(tx, event_id, user_id, assignmentStatusConfirmed)
  return
}
