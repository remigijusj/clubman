package main

import (
  "database/sql"
  "log"
)

type AssignmentRecord struct {
  Id       int
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
    err := rows.Scan(&item.Id, &item.TeamId, &item.UserId, &item.Status)
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
