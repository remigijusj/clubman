package main

import (
  "database/sql"
  "errors"
  "fmt"
  "log"
  "time"
)

// NOTE: delayed, 2 modes based on autoConfirm
func afterAssignmentDelete(event_id, limit_id int) {
  tx, err := db.Begin()
  if err != nil { return }

  err = checkParticipantsLimit(tx, event_id)
  if err != nil { tx.Rollback(); return }

  user_ids, err := listWaitingUsersTx(tx, event_id, limit_id)
  if err != nil { tx.Rollback(); return }

  if len(user_ids) == 0 { tx.Rollback(); return }

  err = updateAssignmentStatusTx(tx, event_id, user_ids[0], assigmentStatusChange())
  if err != nil { tx.Rollback(); return }

  event, err := fetchEventInfoTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  users, err := listUsersByIdTx(tx, user_ids)
  if err != nil { tx.Rollback(); return }

  err = tx.Commit()
  if err != nil { return }

  if autoConfirm {
    notifyEventConfirmed(&event, &users[0])
    event.eachUser(users[1:], notifyEventWaitingUp)
  } else {
    notifyEventToConfirm(&event, &users[0])
    expireAfterGracePeriod(event_id, user_ids[0]) // sleeps
  }
}

func assigmentStatusChange() int {
  if autoConfirm {
    return assignmentStatusConfirmed
  } else {
    return assignmentStatusNotified
  }
}

func checkParticipantsLimit(tx *sql.Tx, event_id int) error {
  max, err := maxTeamUsersTx(tx, event_id)
  if err != nil || max == 0 {
    return errors.New("no maximum")
  }

  cnt, err := countNonWaitingUsersTx(tx, event_id)
  if err != nil {
    return err
  }
  if cnt >= max {
    log.Printf("[APP] CHECK-PARTICIPANTS-LIMIT: %d, %d >= %d\n", event_id, cnt, max)
    return errors.New("event full")
  }
  return nil
}

func listWaitingUsersTx(tx *sql.Tx, event_id, limit_id int) (list []int, err error) {
  list = []int{}

  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-QUEUE: %v, %d, %d\n", err, event_id, limit_id)
    }
  }()

  rows, err := tx.Stmt(query["assignments_queue"]).Query(event_id, assignmentStatusWaiting, limit_id)
  if err != nil { return }
  defer rows.Close()

  var user_id int
  for rows.Next() {
    err = rows.Scan(&user_id)
    if err != nil { return }
    list = append(list, user_id)
  }
  err = rows.Err()

  return
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
    if err != nil {
      log.Printf("[APP] ASSIGNMENT-CONFIRM-STATUS error: %s, %d, %d\n", err, event_id, user_id)
    }
  }()

  status, _, err := fetchAssignmentStatusTx(tx, event_id, user_id)
  if err != nil { return }

  err = checkAssignmentStatusNotified(status)
  if err != nil { return }

  err = updateAssignmentStatusTx(tx, event_id, user_id, assignmentStatusConfirmed)
  return
}

// NOTE: revert to waiting, repeat with next-in-line
func expireAfterGracePeriod(event_id, user_id int) {
  time.Sleep(gracePeriod)

  if limit_id, ok := revertAssignmentToWaiting(event_id, user_id); ok {
    afterAssignmentDelete(event_id, limit_id)
  }
}

func revertAssignmentToWaiting(event_id, user_id int) (int, bool) {
  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] EXPIRE-AFTER-GRACE error: %s, %d, %d\n", err, event_id, user_id)
    }
  }()

  tx, err := db.Begin()
  if err != nil { return 0, false }

  status, id, err := fetchAssignmentStatusTx(tx, event_id, user_id)
  if err != nil { tx.Rollback(); return 0, false }

  err = checkAssignmentStatusNotified(status)
  if err != nil { tx.Rollback(); return 0, false }

  err = updateAssignmentStatusTx(tx, event_id, user_id, assignmentStatusWaiting)
  if err != nil { tx.Rollback(); return 0, false }

  err = tx.Commit()
  if err != nil { return 0, false }

  return id, true
}

func checkAssignmentStatusNotified(status int) error {
  if status != assignmentStatusNotified {
    return errors.New(fmt.Sprintf("invalid status: %d", status))
  }
  return nil
}
