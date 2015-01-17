package main

import (
  "errors"
  "database/sql"
  "log"
)

type AssignmentRecord struct {
  TeamId   int
  UserId   int
  Status   int
}

func listEventAssignments(event_id int) []AssignmentRecord {
  rows, err := query["assignments_event"].Query(event_id)
  return listAssignments(rows, err)
}

func listAssignments(rows *sql.Rows, err error) []AssignmentRecord {
  list := []AssignmentRecord{}
  
  if err != nil {
    log.Printf("[APP] EVENT-ASSIGNMENTS-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item AssignmentRecord
    err := rows.Scan(&item.TeamId, &item.UserId, &item.Status)
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

func createAssignment(event_id, user_id int) error {
  res, err := query["assignment_insert"].Exec(event_id, user_id, 0)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-CREATE error: %s, %d, %d\n", err, event_id, user_id)
    return errors.New("Assignment could not be created")
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return errors.New("Assignment could not be created")
  }
  return nil
}

func cancelAssignment(event_id, user_id int) error {
  res, err := query["assignment_status"].Exec(1, event_id, user_id)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-STATUS error: %s, %d, %d\n", err, event_id, user_id)
    return errors.New("Assignment could not be updated")
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return errors.New("Assignment could not be updated")
  }
  return nil
}

func deleteAssignment(event_id, user_id int) error {
  res, err := query["assignment_delete"].Exec(event_id, user_id)
  if err != nil {
    log.Printf("[APP] ASSIGNMENTS-DELETE error: %s, %d, %d\n", err, event_id, user_id)
    return errors.New("Assignment could not be deleted")
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return errors.New("Assignment could not be deleted")
  }
  return nil
}

func findAssignment(list []AssignmentRecord, user *AuthInfo) *AssignmentRecord {
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
