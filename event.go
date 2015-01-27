package main

import (
  "database/sql"
  "errors"
  "log"
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
    clearAssignments(event_id)
  }
  return nil
}

func cancelEvent(event_id int) error {
  _, err := query["event_status"].Exec(eventStatusCanceled, event_id)
  if err != nil {
    log.Printf("[APP] EVENT-CANCEL error: %s, %d\n", err, event_id)
    return errors.New("Event could not be canceled")
  }
  clearAssignments(event_id)
  return nil
}

func deleteEvent(event_id int) error {
  _, err := query["event_delete"].Exec(event_id)
  if err != nil {
    log.Printf("[APP] EVENT-DELETE error: %s, %d\n", err, event_id)
    return errors.New("Event could not be deleted")
  }
  clearAssignments(event_id)
  return nil
}

func clearEvents(team_id int) error {
  _, err := query["events_clear"].Exec(team_id)
  if err != nil {
    log.Printf("[APP] EVENTS-CLEAR error: %s, %d\n", err, team_id)
  }
  return err
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
