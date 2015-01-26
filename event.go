package main

import (
  "database/sql"
  "errors"
  "log"
  "strconv"
  "time"
)

type EventRecord struct {
  Id       int
  TeamId   int
  StartAt  time.Time
  Minutes  int
  Status   int
}

// NOTE: `date`, `time` just to pull data from event update
type EventForm struct {
  TeamId   int       `form:"team_id"  binding:"required"`
  StartAt  time.Time `form:"start_at"`
  Date     string    `form:"date"     binding:"required"`
  Time     string    `form:"time"     binding:"required"`
  Minutes  int       `form:"minutes"  binding:"required"`
  Status   int       `form:"status"`
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

func (self EventRecord) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func (self EventForm) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func listTeamEvents(team_id int, date_from time.Time) []EventRecord {
  rows, err := query["events_team"].Query(team_id, date_from.Format(dateFormat))
  return listEvents(rows, err)
}

func listEvents(rows *sql.Rows, err error) []EventRecord {
  list := []EventRecord{}

  if err != nil {
    log.Printf("[APP] TEAM-EVENTS-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item EventRecord
    err := rows.Scan(&item.Id, &item.TeamId, &item.StartAt, &item.Minutes, &item.Status)
    if err != nil {
      log.Printf("[APP] TEAM-EVENTS-LIST error: %s\n", err)
    } else {
      // WARNING: we interpret datetimes in DB literally as entered, but all data being UTC
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

func createEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, true, lang)
  if err != nil {
    return 0, err
  }
  cnt := data.eachTime(func(date time.Time) int {
    res, err := query["event_insert"].Exec(team_id, date, data.Minutes, 0)
    if err != nil {
      log.Printf("[APP] EVENTS-CREATE error: %s, %d, %s\n", err, team_id, date)
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
  cnt := data.eachEvent(team_id, func(event_id int) int {
    res, err := query["event_status"].Exec(eventStatusCanceled, event_id)
    if err != nil {
      log.Printf("[APP] EVENTS-CANCEL error: %s, %d\n", err, event_id)
      return 0
    }
    num, _ := res.RowsAffected()
    if num > 0 {
      removeAssignments(event_id)
    }
    return int(num)
  })
  return cnt, nil
}

func removeEvents(team_id int, form *TeamEventsForm, lang string) (int, error) {
  data, err := parseEventsForm(form, false, lang)
  if err != nil {
    return 0, err
  }
  cnt := data.eachEvent(team_id, func(event_id int) int {
    res, err := query["event_delete"].Exec(event_id)
    if err != nil {
      log.Printf("[APP] EVENTS-REMOVE error: %s, %d\n", err, event_id)
      return 0
    }
    num, _ := res.RowsAffected()
    if num > 0 {
      removeAssignments(event_id)
    }
    return int(num)
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
  if need_time && !minutesValid(data.Minutes) {
    return nil, errors.New("Duration must be a positive number, not too big")
  }

  return &data, nil
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

func (self TeamEventsData) eachEvent(team_id int, callback func(int) int) int {
  var kind string
  if self.StartAt.IsZero() {
    kind = "events_by_date"
  } else{
    kind = "events_by_time"
  }

  return self.eachTime(func(date time.Time) int {
    rows, err := query[kind].Query(team_id, date)
    if err != nil {
      log.Printf("[APP] EVENTS-EACH-ASSIGNMENT error: %s, %d, %s\n", err, team_id, date)
      return 0
    }
    defer rows.Close()

    cnt := 0
    var event_id int
    for rows.Next() {
      err := rows.Scan(&event_id)
      if err != nil {
        log.Printf("[APP] EVENTS-EACH-ASSIGNMENT error: %s, %d, %s\n", err, team_id, date)
        continue
      }
      cnt += callback(event_id)
    }
    if err := rows.Err(); err != nil {
      log.Printf("[APP] EVENTS-EACH-ASSIGNMENT error: %s, %d, %s\n", err, team_id, date)
    }

    return cnt
  })
}

func fetchEvent(event_id int) (EventForm, error) {
  var form EventForm
  err := query["event_select"].QueryRow(event_id).Scan(&form.TeamId, &form.StartAt, &form.Minutes, &form.Status)
  if err != nil {
    log.Printf("[APP] EVENT-SELECT error: %s, %#v\n", err, form)
    err = errors.New("Event was not found")
  }
  // WARNING: see the comment in listEvents
  form.StartAt = form.StartAt.UTC()
  return form, err
}

func updateEvent(event_id int, form *EventForm, lang string) error {
  err := parseEventForm(form, lang)
  if err != nil {
    return err
  }
  _, err = query["event_update"].Exec(form.TeamId, form.StartAt, form.Minutes, form.Status, event_id)
  if err != nil {
    log.Printf("[APP] EVENT-UPDATE error: %s, %d\n", err, event_id)
    return errors.New("Event could not be updated")
  }
  if form.Status == eventStatusCanceled {
    removeAssignments(event_id)
  }
  return nil
}

func cancelEvent(event_id int) error {
  _, err := query["event_status"].Exec(eventStatusCanceled, event_id)
  if err != nil {
    log.Printf("[APP] EVENT-CANCEL error: %s, %d\n", err, event_id)
    return errors.New("Event could not be canceled")
  }
  removeAssignments(event_id)
  return nil
}

func deleteEvent(event_id int) error {
  _, err := query["event_delete"].Exec(event_id)
  if err != nil {
    log.Printf("[APP] EVENT-DELETE error: %s, %d\n", err, event_id)
    return errors.New("Event could not be deleted")
  }
  removeAssignments(event_id)
  return nil
}

func parseEventForm(form *EventForm, lang string) error {
  date, err := time.Parse(dateFormats[lang], form.Date)
  if err != nil {
    return errors.New("Date must be valid")
  }
  when, err := time.Parse(timeFormat, form.Time)
  if err != nil {
    return errors.New("Start time has invalid format")
  }
  hour := hourDuration(when)
  form.StartAt = date.Add(hour)

  if !minutesValid(form.Minutes) {
    return errors.New("Duration must be a positive number, not too big")
  }
  if !(form.Status == -2 || form.Status == 0) {
    return errors.New("Status is invalid")
  }
  return nil
}

func minutesValid(minutes int) bool {
  return minutes > 0 && minutes < 6 * 60
}

func collectEventIds(list []EventRecord) []int {
  event_ids := make([]int, len(list))
  for i, item := range list {
    event_ids[i] = item.Id
  }
  return event_ids
}
