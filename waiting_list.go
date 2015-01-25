package main

import (
  "database/sql"
  "errors"
  "log"
)

// NOTE: delayed
func afterAssignmentDelete(event_id int, lang string) {
  tx, err := db.Begin()
  if err != nil { return }

  user_id, err := firstWaitingUserId(tx, event_id)
  if err != nil { tx.Rollback(); return }

  err = updateAssignmentStatusTx(tx, event_id, user_id, assignmentStatusNotified)
  if err != nil { tx.Rollback(); return }

  data, err := fetchUserProfile(user_id)
  if err != nil { tx.Rollback(); return }

  err = tx.Commit()
  if err != nil { return }

  sendWaitingConfirmationEmail(lang, data.Email, user_id, event_id)
}

func firstWaitingUserId(tx *sql.Tx, event_id int) (int, error) {
  max, err := maxTeamUsersTx(tx, event_id)
  if err != nil {
    return 0, err
  }
  counts, err := groupedAssignmentsStatusTx(tx, event_id)
  if err != nil {
    return 0, err
  }
  ok := firstWaitingUserCheck(counts, max)
  if !ok {
    log.Printf("[APP] ASSIGNMENTS-FIRST-WAITING condition: %d\n", event_id)
    return 0, errors.New("")
  }
  return firstAssignmentByStatusTx(tx, event_id, assignmentStatusWaiting)
}

func firstWaitingUserCheck(counts map[int]int, max int) bool {
  return counts[assignmentStatusConfirmed] < max &&
    counts[assignmentStatusNotified] == 0 &&
    counts[assignmentStatusWaiting] > 0
}

func groupedAssignmentsStatusTx(tx *sql.Tx, event_id int) (counts map[int]int, err error) {
  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-CHECK-STATUS error: %s, %d\n", err, event_id)
    }
  }()
  counts = make(map[int]int, 3)

  rows, err := tx.Stmt(query["assignments_check"]).Query(event_id)
  if err != nil { return }
  defer rows.Close()

  var status, count int
  for rows.Next() {
    err = rows.Scan(&status, &count)
    if err != nil { return }
    counts[status] = count
  }
  err = rows.Err()
  return
}

func firstAssignmentByStatusTx(tx *sql.Tx, event_id, status int) (int, error) {
  var user_id int
  err := tx.Stmt(query["assignments_first"]).QueryRow(event_id, status).Scan(&user_id)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-STATUS-FIRST error: %s, %d, %d\n", err, event_id, status)
  }
  return user_id, err
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
