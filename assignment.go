package main

import (
  "errors"
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

func listUserAssignments(user_id int) []UserAssignment {
  rows, err := query["assignments_user"].Query(user_id)
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
  res, err := query["assignment_status"].Exec(-2, event_id, user_id)
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

func mapAssignedPeriod(user_id int, from, till time.Time) map[string]bool {
  data := map[string]bool{}
  var when time.Time
  rows, err := query["assignments_period"].Query(user_id, from.Format(dateFormat), till.Format(dateFormat))
  if err != nil {
    log.Printf("[APP] USER-ASSIGNMENTS-PERIOD error: %s\n", err)
    return data
  }
  defer rows.Close()
  for rows.Next() {
    err := rows.Scan(&when)
    if err != nil {
      log.Printf("[APP] USER-ASSIGNMENTS-PERIOD error: %s\n", err)
    } else {
      data[when.UTC().Format(dateFormat)] = true
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-ASSIGNMENTS-PERIOD error: %s\n", err)
  }
  return data
}
