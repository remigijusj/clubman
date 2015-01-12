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
  rows, err := query["events_team"].Query(team_id)
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
      // NOTE: we interpret datetimes in DB literally as entered, but all data being UTC
      //   time.Parse gives UTC already, but rows.Scan gives us local times (why?), so convert!
      item.StartAt = item.StartAt.UTC()
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] TEAM-EVENTS-LIST error: %s\n", err)
  }
  return list
}

func addEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, true, lang)
  if err != nil {
    return 0, err
  }
  cnt := data.iterate(func(date time.Time) bool {
    res, err := query["events_create"].Exec(team_id, date, data.Minutes, 0)
    if err != nil {
      log.Printf("[APP] EVENTS-ADD error: %s, %d, %s\n", err, team_id, date)
      return false
    }
    num, err := res.RowsAffected()
    return num > 0 && err == nil
  })
  return cnt, nil
}

func cancelEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil {
    return 0, err
  }
  cnt := data.iterate(func(date time.Time) bool {
    var kind string
    if data.StartAt.IsZero() {
      kind = "events_status_date"
    } else{
      kind = "events_status_time"
    }
    res, err := query[kind].Exec(1, team_id, date)
    if err != nil {
      log.Printf("[APP] EVENTS-CANCEL error: %s, %d, %s\n", err, team_id, date)
      return false
    }
    num, err := res.RowsAffected()
    return num > 0 && err == nil
  })
  return cnt, nil
}

func removeEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil {
    return 0, err
  }
  cnt := data.iterate(func(date time.Time) bool {
    var kind string
    if data.StartAt.IsZero() {
      kind = "events_delete_date"
    } else{
      kind = "events_delete_time"
    }
    res, err := query[kind].Exec(team_id, date)
    if err != nil {
      log.Printf("[APP] EVENTS-DELETE error: %s, %d, %s\n", err, team_id, date)
      return false
    }
    num, err := res.RowsAffected()
    return num > 0 && err == nil
  })
  return cnt, nil
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
  if need_time && data.Minutes <= 0 || data.Minutes >= 6 * 60 {
    return nil, errors.New("Duration must be a positive number, not too big")
  }

  return &data, nil
}

func (self TeamEventsData) iterate(callback func(time.Time) bool) int {
  hour := self.StartAt.Sub(self.StartAt.Truncate(24 * time.Hour))
  date := self.DateFrom.Add(hour)
  ceil := self.DateTill.AddDate(0, 0, 1)
  cnt := 0
  for ; date.Before(ceil); date = date.AddDate(0, 0, 1) {
    wday := int(date.Weekday())
    if self.Weekdays[wday] {
      if callback(date) {
        cnt++
      }
    }
  }
  return cnt
}
