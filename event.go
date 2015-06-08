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

type EventInfo struct {
  Id       int
  Name     string
  StartAt  time.Time
  Minutes  int
  Status   int
}

func listTeamEvents(team_id int, date_from time.Time) []EventRecord {
  rows, err := query["events_team"].Query(team_id, date_from.Format(dateFormat))
  return listEvents(rows, err)
}

func listEvents(rows *sql.Rows, err error) (list []EventRecord) {
  list = []EventRecord{}

  defer func() {
    if err != nil {
      log.Printf("[APP] LIST-EVENTS error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  for rows.Next() {
    var item EventRecord
    err = rows.Scan(&item.Id, &item.TeamId, &item.StartAt, &item.Minutes, &item.Status)
    if err != nil { return }
    // WARNING: we interpret datetimes in DB literally as entered, but all data being UTC
    //   time.Parse gives UTC already, but rows.Scan gives us local times (why?), so convert!
    item.StartAt = item.StartAt.UTC()
    list = append(list, item)
  }
  err = rows.Err()

  return
}

func listEventsUnderLimit(time_from, time_till time.Time) []int {
  rows, err := query["events_under"].Query(time_from.Format(fullFormat), time_till.Format(fullFormat))
  return listEventsIds(rows, err)
}

// NOTE: compare with listEvents
func listEventsIds(rows *sql.Rows, err error) (list []int) {
  list = []int{}
  defer func() {
    if err != nil {
      log.Printf("[APP] LIST-EVENTS-IDS error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  var id int
  for rows.Next() {
    err = rows.Scan(&id)
    if err != nil { return }
    list = append(list, id)
  }
  err = rows.Err()

  return
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

func fetchEventInfoTx(tx *sql.Tx, event_id int) (EventInfo, error) {
  var event EventInfo
  err := tx.Stmt(query["event_select_info"]).QueryRow(event_id).Scan(&event.Id, &event.Name, &event.StartAt, &event.Minutes, &event.Status)
  if err != nil {
    log.Printf("[APP] FETCH-EVENT-INFO: %v, %d\n", err, event_id)
  } else {
    event.StartAt = event.StartAt.UTC()
  }
  return event, err
}

func updateEvent(event_id int, form *EventForm, lang string) error {
  err := parseEventForm(form, lang)
  if err != nil { return err }

  clear := form.Status == eventStatusCanceled
  return performEventAction(event_id, updateEventRecordTx, clear, form)
}

func cancelEvent(event_id int) error {
  return performEventAction(event_id, cancelEventRecordTx, true, nil)
}

func deleteEvent(event_id int) error {
  return performEventAction(event_id, deleteEventRecordTx, true, nil)
}

func performEventAction(event_id int, action (func(*sql.Tx, int, *EventForm) error), clear bool, form *EventForm) (err error) {
  defer func() {
    if err != nil {
      log.Printf("[APP] EVENT-ACTION error: %s, %d\n", err, event_id)
    }
  }()

  tx, err := db.Begin()
  if err != nil { return }

  event, err := fetchEventInfoTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  users, err := listUsersOfEventTx(tx, event_id)
  if err != nil { tx.Rollback(); return }

  err = action(tx, event_id, form)
  if err != nil { tx.Rollback(); return }

  err = tx.Commit()
  if err != nil { return }

  // after-callbacks
  if clear {
    clearAssignments(event_id)
  }
  if !event.isPast() {
    if clear {
      go event.eachUser(users, notifyEventCancel)
    } else if form != nil { // also: !form.StartAt.Equal(event.StartAt)
      go event.eachUser(users, notifyEventUpdate)
    }
  }

  return
}

func updateEventRecordTx(tx *sql.Tx, event_id int, form *EventForm) error {
  _, err := tx.Stmt(query["event_update"]).Exec(form.TeamId, form.StartAt, form.Minutes, form.Status, event_id)
  if err != nil {
    log.Printf("[APP] EVENT-UPDATE-TX: error %v, %d, %v\n", err, event_id, *form)
    err = errors.New("Event could not be updated")
  }
  return err
}

func cancelEventRecordTx(tx *sql.Tx, event_id int, form *EventForm) error {
  _, err := tx.Stmt(query["event_status"]).Exec(eventStatusCanceled, event_id)
  if err != nil {
    log.Printf("[APP] EVENT-CANCEL-TX: error %v, %d\n", err, event_id)
    err = errors.New("Event could not be canceled")
  }
  return err
}

func deleteEventRecordTx(tx *sql.Tx, event_id int, form *EventForm) error {
  _, err := tx.Stmt(query["event_delete"]).Exec(event_id)
  if err != nil {
    log.Printf("[APP] EVENT-DELETE-TX: error %v, %d\n", err, event_id)
    err = errors.New("Event could not be deleted")
  }
  return err
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

// NOTE: delayed, cron, events of tomorrow
func autoCancelEvents() {
  when := time.Now() // WARNING: need localtime, really
  from := when.Add(cancelAhead)
  till := from.Add(1 * time.Hour)

  list := listEventsUnderLimit(from, till)
  log.Printf("[APP] AUTOCANCEL-EVENT-INFO: %s, %s, %v\n", from.Format(fullFormat), till.Format(fullFormat), list)
  for _, event_id := range list {
    cancelEvent(event_id)
  }
}

// NOTE: should be in notifications.go
func notifyEventParticipants(sender_id, event_id int, subject, message string) int {
  var count int
  rows, err := query["users_of_event"].Query(event_id)
  users := listUsersContact(rows, err)
  var from string
  if sender, err := fetchUserProfile(sender_id); err == nil {
    from = sender.Email
  }
  for _, user := range users {
    if ok := sendEmail(user.Email, subject, message, from); ok {
      count++
    }
  }
  return count
}
