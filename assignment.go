package main

import (
  "database/sql"
  "log"
  "time"
)

type EventAssignment struct {
  UserId      int
  UserName    string
  Status      int
}

type UserAssignment struct {
  EventId     int
  TeamName    string
  StartAt     time.Time
  Minutes     int
  EventStatus int
  Status      int
}

func (self UserAssignment) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func listEventAssignments(event_id int) (list []EventAssignment) {
  list = []EventAssignment{}

  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-EVENT error: %s, %d\n", err, event_id)
    }
  }()

  rows, err := query["assignments_event"].Query(event_id)
  if err != nil { return }
  defer rows.Close()

  for rows.Next() {
    var item EventAssignment
    err = rows.Scan(&item.UserId, &item.UserName, &item.Status)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
}

func listUserAssignments(user_id int, date_from time.Time) (list []UserAssignment) {
  list = []UserAssignment{}

  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-USER error: %s, %d, %v\n", err, user_id, date_from)
    }
  }()

  rows, err := query["assignments_user"].Query(user_id, date_from.Format(dateFormat))
  if err != nil { return }
  defer rows.Close()

  for rows.Next() {
    var item UserAssignment
    err = rows.Scan(&item.EventId, &item.TeamName, &item.StartAt, &item.Minutes, &item.EventStatus, &item.Status)
    if err != nil { return }
    // WARNING: see the comment in listEvents  
    item.StartAt = item.StartAt.UTC()
    list = append(list, item)
  }
  err = rows.Err()

  return
}

func findAssignment(list []EventAssignment, user *AuthInfo) *EventAssignment {
  if user == nil {
    return nil
  }
  for _, item := range list {
    if item.UserId == user.Id {
      return &item
    }
  }
  return nil
}

func fetchAssignmentStatusTx(tx *sql.Tx, event_id, user_id int) (int, int, error) {
  var status, id int
  err := tx.Stmt(query["assignment_status"]).QueryRow(event_id, user_id).Scan(&id, &status)
  if err != nil {
    log.Printf("[APP] ASSIGNMENT-STATUS error: %s, %d, %d\n", err, event_id, user_id)
  }
  return status, id, err
}

func createAssignmentTx(tx *sql.Tx, event_id, user_id, status int) error {
  res, err := tx.Stmt(query["assignment_insert"]).Exec(event_id, user_id, status)
  if err != nil {
    log.Printf("[APP] ASSIGNMENT-INSERT error: %s, %d, %d\n", err, event_id, user_id)
    return err
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return err
  }
  return nil
}

func deleteAssignmentTx(tx *sql.Tx, event_id, user_id int) error {
  res, err := tx.Stmt(query["assignment_delete"]).Exec(event_id, user_id)
  if err != nil {
    log.Printf("[APP] ASSIGNMENT-DELETE error: %s, %d, %d\n", err, event_id, user_id)
    return err
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return err
  }
  return nil
}

func mapAssignedStatus(event_ids []int, user_id int) (data map[int]int) {
  data = make(map[int]int, len(event_ids))

  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-STATUS error: %s, %d, %v\n", err, user_id, event_ids)
    }
  }()

  if len(event_ids) == 0 { return }

  rows, err := multiQuery("assignments_status", event_ids, user_id)
  if err != nil { return }
  defer rows.Close()

  var event_id, status int
  for rows.Next() {
    err = rows.Scan(&event_id, &status)
    if err != nil { return }
    data[event_id] = status
  }
  err = rows.Err()

  return
}

func mapParticipantCounts(event_ids []int) (data map[int]int) {
  data = make(map[int]int, len(event_ids))

  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-COUNTS error: %s, %v\n", err, event_ids)
    }
  }()

  if len(event_ids) == 0 { return }

  rows, err := multiQuery("assignments_counts", event_ids)
  if err != nil { return }
  defer rows.Close()

  var event_id, count int
  for rows.Next() {
    err = rows.Scan(&event_id, &count)
    if err != nil { return }
    data[event_id] = count
  }
  err = rows.Err()

  return
}

func countAssignmentsTx(tx *sql.Tx, event_id int) (int, error) {
  var count int
  err := tx.Stmt(query["assignments_count"]).QueryRow(event_id).Scan(&count)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-COUNT error: %s, %d\n", err, event_id)
  }
  return count, err
}

func clearAssignments(event_ids ...int) error {
  if len(event_ids) == 0 {
    return nil
  }
  _, err := multiExec("assignments_clear", event_ids)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-CLEAR error: %s, %v\n", err, event_ids)
  }
  return err
}

func notifyAssignmentAction(event_id, user_id, status int) {
  tx, err := db.Begin()
  if err != nil { return }

  user, err := fetchUserContactTx(tx, user_id)
  if err != nil { tx.Rollback(); return }

  event, err := fetchEventInfoTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  err = tx.Commit()
  if err != nil { return }

  if status != 0 {
    confirmed := status == assignmentStatusConfirmed
    sendAssignmentCreatedEmail(user.Email, user.Language, &event, confirmed)
  } else {
    sendAssignmentDeletedEmail(user.Email, user.Language, &event)
  }
}
