package main

import (
  "log"
  "time"
)

type EventRecord struct {
  Id       int
  StartAt  time.Time
  Minutes  int
  Status   int
}

type TeamEventsForm struct {
  DateFrom string `form:"date_from" binding:"required"`
  DateTill string `form:"date_till" binding:"required"`
  Weekdays string `form:"weekdays"`
  StartAt  string `form:"start_at"`
  Minutes  string `form:"minutes"`
}

func listTeamEvents(team_id int) []EventRecord {
  list := []EventRecord{}
  rows, err := query["team_events"].Query(team_id)
  if err != nil {
    log.Printf("[APP] TEAM-EVENTS-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item EventRecord
    err := rows.Scan(&item.Id, &item.StartAt, &item.Minutes, &item.Status)
    if err != nil {
      log.Printf("[APP] TEAM-EVENTS-LIST error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] TEAM-EVENTS-LIST error: %s\n", err)
  }
  return list
}
