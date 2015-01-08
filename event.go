package main

import (
  "errors"
  "log"
  "strconv"
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
  Weekdays []int  `form:"weekdays"`
  StartAt  string `form:"start_at"`
  Minutes  string `form:"minutes"`
}

type TeamEventsData struct {
  DateFrom time.Time
  DateTill time.Time
  Weekdays []bool
  StartAt  time.Time
  Minutes  int
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

func addEvents(team_id int, form *TeamEventsForm) (int, error) {
  data, err := parseEventsForm(form, true)
  if err != nil {
    return 0, err
  }
  log.Printf("=> EVENTS-ADD: %#v\n", *data)
  // <<< query
  return 0, nil
}

func cancelEvents(team_id int, form *TeamEventsForm) (int, error) {
  data, err := parseEventsForm(form, false)
  if err != nil {
    return 0, err
  }
  log.Printf("=> EVENTS-CANCEL: %#v\n", *data)
  // <<< query
  return 0, nil
}

func removeEvents(team_id int, form *TeamEventsForm) (int, error) {
  data, err := parseEventsForm(form, false)
  if err != nil {
    return 0, err
  }
  log.Printf("=> EVENTS-REMOVE: %#v\n", *data)
  // <<< query
  return 0, nil
}

func parseEventsForm(form *TeamEventsForm, need_time bool) (*TeamEventsData, error) {
  var data TeamEventsData

  var err1, err2 error
  data.DateFrom, err1 = time.Parse(dateFormat, form.DateFrom)
  data.DateTill, err2 = time.Parse(dateFormat, form.DateTill)
  if err1 != nil || err2 != nil || data.DateTill.Before(data.DateFrom) {
    return nil, errors.New("Dates must be valid")
  }

  data.Weekdays = make([]bool, 7)
  if len(form.Weekdays) > 0 {
    for _, val := range form.Weekdays {
      if val >= 0 && val < 7 {
        data.Weekdays[val] = true
      }
    }
  } else {
    for i, _ := range data.Weekdays {
      data.Weekdays[i] = true
    }
  }

  data.StartAt, err1 = time.Parse(timeFormat, form.StartAt)
  if need_time && err1 != nil {
    return nil, errors.New("Start time has invalid format")
  }

  data.Minutes, _ = strconv.Atoi(form.Minutes)
  if need_time && data.Minutes <= 0 || data.Minutes >= 5 * 60 {
    return nil, errors.New("Duration must be a positive number, not too big")
  }

  return &data, nil
}
