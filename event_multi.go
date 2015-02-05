package main

import (
  "database/sql"
  "errors"
  "fmt"
  "log"
  "strconv"
  "strings"
  "time"
)

type TeamEventsForm struct {
  DateFrom string `form:"date_from" binding:"required"`
  DateTill string `form:"date_till" binding:"required"`
  Weekdays []int  `form:"weekdays"`
  OnlyAt   string `form:"only_at"`
  StartAt  string `form:"start_at"`
  Minutes  string `form:"minutes"`
  Status   string `form:"status"`
}

type TeamEventsData struct {
  DateFrom time.Time
  DateTill time.Time
  Weekdays []bool
  OnlyAt   time.Time
  StartAt  time.Time
  Minutes  int
  Status   int
}

func (self TeamEventsData) eachTime(callback func(time.Time) int) int {
  hour := hourDuration(self.StartAt)
  date := self.DateFrom.Add(hour)
  ceil := self.DateTill.AddDate(0, 0, 1)
  cnt := 0
  for ; date.Before(ceil); date = date.AddDate(0, 0, 1) {
    wday := int(date.Weekday())
    if self.Weekdays[wday] {
      cnt += callback(date)
    }
  }
  return cnt
}

func (self TeamEventsData) eventIds(team_id int) (list []int, near bool) {
  list = []int{}

  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] EVENTS-MULTI error: %s, %d, %v\n", err, team_id, self)
    }
  }()

  rows, err := query["events_multi"].Query(team_id, self.DateFrom.Format(dateFormat), self.DateTill.AddDate(0, 0, 1).Format(dateFormat))
  if err != nil { return }
  defer rows.Close()

  var event_id int
  var start_at time.Time
  for rows.Next() {
    err := rows.Scan(&event_id, &start_at)
    if err != nil { return }
    // WARNING: see the comment in listEvents
    start_at = start_at.UTC()
    if self.matchTime(start_at) {
      list = append(list, event_id)
      near = near || isNear(start_at) // at least 1 event near enough for sms
    }
  }
  err = rows.Err()

  return
}

func (self TeamEventsData) matchTime(t time.Time) bool {
  wday := int(t.Weekday())
  if !self.Weekdays[wday] {
    return false
  }
  s := self.OnlyAt
  if s.IsZero() {
    return true
  } else {
    return s.Hour() == t.Hour() && s.Minute() == t.Minute()
  }
}

func createEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, true, lang)
  if err != nil {
    return 0, err
  }
  cnt := data.eachTime(func(date time.Time) int {
    res, err := query["event_insert"].Exec(team_id, date, data.Minutes, eventStatusActive)
    if err != nil {
      log.Printf("[APP] EVENT-INSERT error: %s, %d, %s, %d\n", err, team_id, date, data.Minutes)
      return 0
    }
    num, _ := res.RowsAffected()
    return int(num)
  })
  return cnt, nil
}

func updateEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil { return 0, err }

  list, near := data.eventIds(team_id)
  if len(list) == 0 { return 0, nil }

  users, err := listUsersOfEvents(list)
  if err != nil { return 0, nil }

  team, err := fetchTeam(team_id)
  if err != nil { return 0, nil }

  res, err := updateEventsRecords(list, data)
  if err != nil { return 0, nil }

  num, _ := res.RowsAffected()
  if num > 0 {
    if !data.isPast() {
      if data.Status == eventStatusCanceled {
        go data.eachUser(users, notifyEventMultiCancel, &team, near)
      } else {
        go data.eachUser(users, notifyEventMultiUpdate, &team, near)
      }
    }
  }
  return int(num), nil
}

// NOTE: building the query in-place
func updateEventsRecords(event_ids []int, data *TeamEventsData) (sql.Result, error) {
  cond := make([]string, 0, 3)

  if !data.OnlyAt.IsZero() && !data.StartAt.IsZero() {
    part := fmt.Sprintf("start_at=datetime(start_at, 'start of day', '%d hours', '%d minutes')", data.StartAt.Hour(), data.StartAt.Minute())
    cond = append(cond, part)
  }

  if data.Minutes > 0 {
    part := fmt.Sprintf("minutes=%d", data.Minutes)
    cond = append(cond, part)
  }

  if true {
    part := fmt.Sprintf("status=%d", data.Status)
    cond = append(cond, part)
  }

  qry := fmt.Sprintf("UPDATE events SET %s WHERE id IN (?)", strings.Join(cond, ", "))
  qry, list := multi(qry, event_ids)
  res, err := db.Exec(qry, list...)
  if err != nil {
    log.Printf("[APP] EVENT-UPDATE-MULTI error: %s, %v, %v\n", err, data, event_ids)
  }
  return res, err
}

func deleteEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil { return 0, err }

  list, near := data.eventIds(team_id)
  if len(list) == 0 { return 0, nil }

  users, err := listUsersOfEvents(list)
  if err != nil { return 0, nil }

  team, err := fetchTeam(team_id)
  if err != nil { return 0, nil }

  res, err := deleteEventsRecords(list)
  if err != nil { return 0, nil }

  num, _ := res.RowsAffected()
  if num > 0 {
    clearAssignments(list...)
    if !data.isPast() {
      go data.eachUser(users, notifyEventMultiCancel, &team, near)
    }
  }
  return int(num), nil
}

func deleteEventsRecords(event_ids []int) (sql.Result, error) {
  res, err := multiExec("event_delete", event_ids)
  if err != nil {
    log.Printf("[APP] EVENT-DELETE-MULTI error: %s, %v\n", err, event_ids)
  }
  return res, err
}

func parseEventsForm(form *TeamEventsForm, create bool, lang string) (*TeamEventsData, error) {
  var data TeamEventsData

  var err1, err2 error
  data.DateFrom, err1 = time.Parse(dateFormats[lang], form.DateFrom)
  data.DateTill, err2 = time.Parse(dateFormats[lang], form.DateTill)
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

  data.OnlyAt, _ = time.Parse(timeFormat, form.OnlyAt)

  if create && form.StartAt == "" {
    return nil, errors.New("Please enter start time")
  }
  data.StartAt, err1 = time.Parse(timeFormat, form.StartAt)
  if create && err1 != nil {
    return nil, errors.New("Start time has invalid format")
  }
  if !create && data.OnlyAt.IsZero() && !data.StartAt.IsZero() {
    return nil, errors.New("To update Start time, you must also enter Filter time")
  }

  data.Minutes, _ = strconv.Atoi(form.Minutes)
  if create && !minutesValid(data.Minutes) {
    return nil, errors.New("Duration must be a positive number, not too big")
  }

  data.Status, _ = strconv.Atoi(form.Status)
  if data.Status != eventStatusActive && data.Status != eventStatusCanceled {
    return nil, errors.New("Status is invalid")
  }

  return &data, nil
}
