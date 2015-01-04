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
