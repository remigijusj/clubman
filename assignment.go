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

func listEventAssignments(event_id int) []EventAssignment {
  rows, err := query["assignments_event"].Query(event_id)
  list := []EventAssignment{}
  
  if err != nil {
    log.Printf("[APP] EVENT-ASSIGNMENTS-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item EventAssignment
    err := rows.Scan(&item.UserId, &item.UserName, &item.Status)
    if err != nil {
      log.Printf("[APP] EVENT-ASSIGNMENTS-LIST error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] EVENT-ASSIGNMENTS-LIST error: %s\n", err)
  }
  return list
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

func listUserAssignments(user_id int, date_from time.Time) []UserAssignment {
  rows, err := query["assignments_user"].Query(user_id, date_from.Format(dateFormat))
  list := []UserAssignment{}

  if err != nil {
    log.Printf("[APP] USER-ASSIGNMENTS-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item UserAssignment
    err := rows.Scan(&item.EventId, &item.TeamName, &item.StartAt, &item.Minutes, &item.EventStatus, &item.Status)
    if err != nil {
      log.Printf("[APP] USER-ASSIGNMENTS-LIST error: %s\n", err)
    } else {
      // WARNING: see the comment in listEvents  
      item.StartAt = item.StartAt.UTC()
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-ASSIGNMENTS-LIST error: %s\n", err)
  }
  return list
}

func createAssignmentTx(tx *sql.Tx, event_id, user_id, status int) error {
  res, err := tx.Stmt(query["assignment_insert"]).Exec(event_id, user_id, status)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-CREATE error: %s, %d, %d\n", err, event_id, user_id)
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
    log.Printf("[APP] ASSIGNMENTS-DELETE error: %s, %d, %d\n", err, event_id, user_id)
    return err
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return err
  }
  return nil
}

func mapAssignedStatus(event_ids []int, user_id int) map[int]int {
  data := make(map[int]int, len(event_ids))

  rows, err := queryMultiple("assignments_status", event_ids, user_id)
  if err != nil {
    log.Printf("[APP] USER-ASSIGNMENTS-PERIOD error: %s\n", err)
    return data
  }
  defer rows.Close()

  var event_id, status int
  for rows.Next() {
    err := rows.Scan(&event_id, &status)
    if err != nil {
      log.Printf("[APP] USER-ASSIGNMENTS-PERIOD error: %s\n", err)
    } else {
      data[event_id] = status
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-ASSIGNMENTS-PERIOD error: %s\n", err)
  }

  return data
}

func mapParticipantCounts(event_ids []int) map[int]int {
  data := make(map[int]int, len(event_ids))
  if len(event_ids) == 0 {
    return data
  }

  rows, err := queryMultiple("assignments_counts", event_ids)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-COUNT error: %s\n", err)
    return data
  }
  defer rows.Close()

  var event_id, count int
  for rows.Next() {
    err := rows.Scan(&event_id, &count)
    if err != nil {
      log.Printf("[APP] ASSIGNMENTS-COUNT error: %s\n", err)
    } else {
      data[event_id] = count
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] ASSIGNMENTS-COUNT error: %s\n", err)
  }

  return data
}

func countAssignmentsTx(tx *sql.Tx, event_id int) (int, error) {
  var count int
  err := tx.Stmt(query["assignments_count"]).QueryRow(event_id).Scan(&count)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-COUNT-EVENT error: %s, %d\n", err, event_id)
  }
  return count, err
}
