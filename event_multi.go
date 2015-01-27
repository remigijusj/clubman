package main

import (
  "errors"
  "log"
  "strconv"
  "time"
)

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

func (self TeamEventsData) eventIds(team_id int) (list []int) {
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
    if self.matchTime(start_at.UTC()) {
      list = append(list, event_id)
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
  s := self.StartAt
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
    res, err := query["event_insert"].Exec(team_id, date, data.Minutes, 0)
    if err != nil {
      log.Printf("[APP] EVENT-INSERT error: %s, %d, %s, %d\n", err, team_id, date, data.Minutes)
      return 0
    }
    num, _ := res.RowsAffected()
    return int(num)
  })
  return cnt, nil
}

func cancelEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil {
    return 0, err
  }
  list := data.eventIds(team_id)
  if len(list) == 0 {
    return 0, nil
  }
  res, err := multiExec("event_status", eventStatusCanceled, list)
  if err != nil {
    log.Printf("[APP] EVENT-STATUS error: %s, %v\n", err, list)
    return 0, nil
  }
  num, _ := res.RowsAffected()
  if num > 0 {
    clearAssignments(list...)
  }
  return int(num), nil
}

func removeEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil {
    return 0, err
  }
  list := data.eventIds(team_id)
  if len(list) == 0 {
    return 0, nil
  }
  res, err := multiExec("event_delete", list)
  if err != nil {
    log.Printf("[APP] EVENT-DELETE error: %s, %v\n", err, list)
    return 0, nil
  }
  num, _ := res.RowsAffected()
  if num > 0 {
    clearAssignments(list...)
  }
  return int(num), nil
}

func parseEventsForm(form *TeamEventsForm, need_time bool, lang string) (*TeamEventsData, error) {
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

  if need_time && form.StartAt == "" {
    return nil, errors.New("Please enter start time")
  }
  data.StartAt, err1 = time.Parse(timeFormat, form.StartAt)
  if need_time && err1 != nil {
    return nil, errors.New("Start time has invalid format")
  }

  data.Minutes, _ = strconv.Atoi(form.Minutes)
  if need_time && !minutesValid(data.Minutes) {
    return nil, errors.New("Duration must be a positive number, not too big")
  }

  return &data, nil
}
