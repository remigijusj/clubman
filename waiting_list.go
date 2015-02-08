package main

import (
  "database/sql"
  "errors"
  "fmt"
  "log"
  "time"
)

type EventQueueItem struct {
  Id          int
  UserId      int
  Status      int
}

// NOTE: delayed, 2 modes based on autoConfirm
func afterAssignmentDelete(event_id, limit_id int) {
  tx, err := db.Begin()
  if err != nil { return }

  max, err := maxTeamUsersTx(tx, event_id)
  if err != nil || max == 0 { tx.Rollback(); return }

  list, err := listEventQueueItemsTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  user_ids, lucky_cnt := shiftWaitingUsers(list, max, limit_id)
  if len(user_ids) == 0 { tx.Rollback(); return }

  err = updateAssignmentStatusTx(tx, event_id, assigmentStatusChange(), user_ids[0:lucky_cnt]...)
  if err != nil { tx.Rollback(); return }

  event, err := fetchEventInfoTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  users, err := listUsersByIdTx(tx, user_ids)
  if err != nil { tx.Rollback(); return }

  err = tx.Commit()
  if err != nil { return }

  if autoConfirm {
    notifyWaitingList(&event, users, lucky_cnt)
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

func listEventQueueItemsTx(tx *sql.Tx, event_id int) (list []EventQueueItem, err error) {
  list = []EventQueueItem{}

  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-QUEUE: %v, %d\n", err, event_id)
    }
  }()

  rows, err := tx.Stmt(query["assignments_queue"]).Query(event_id)
  if err != nil { return }
  defer rows.Close()

  for rows.Next() {
    var item EventQueueItem
    err = rows.Scan(&item.Id, &item.UserId, &item.Status)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
}

func shiftWaitingUsers(list []EventQueueItem, max, limit_id int) ([]int, int) {
  return nil, 0 // <<< TODO, if autoConfirm: [user_id], 1
}

func updateAssignmentStatusTx(tx *sql.Tx, event_id, status int, user_ids ...int) error {
  qry, list := multi(queries["assignment_update"], status, event_id, user_ids)
  res, err := tx.Exec(qry, list...)
  if err != nil {
    log.Printf("[APP] ASSIGNMENT-UPDATE-STATUS error: %s, %d, %d\n", err, event_id, user_ids)
    return err
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return err
  }
  return nil
}

func notifyWaitingList(event *EventInfo, users []UserContact, lucky_cnt int) {
  for i, user := range users {
    if i < lucky_cnt {
      notifyEventConfirmed(event, &user)
    } else {
      notifyEventWaitingUp(event, &user, i-lucky_cnt+1) // waiting num
    }
  }
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

  err = updateAssignmentStatusTx(tx, event_id, assignmentStatusConfirmed, user_id)
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

  err = updateAssignmentStatusTx(tx, event_id, assignmentStatusWaiting, user_id)
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
