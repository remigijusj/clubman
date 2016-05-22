package main

import (
  "time"
)

func (self EventRecord) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func (self EventForm) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func (self EventInfo) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func (self TeamEventsData) FinishAt() time.Time {
  return self.StartAt.Add(time.Duration(self.Minutes) * time.Minute)
}

func (self *TeamEventsData) isPast() bool {
  return self.DateTill.Before(today())
}

func (self *EventInfo) isPast() bool {
  return self.StartAt.Before(today())
}

func (self *TeamEventsData) eachUser(users []UserContact, action (func (*TeamEventsData, *UserContact, *TeamForm, bool)), team *TeamForm, near bool) {
  for _, user := range users {
    action(self, &user, team, near)
  }
}

func (self *EventInfo) eachUser(users []UserContact, action (func (*EventInfo, *UserContact))) {
  for _, user := range users {
    action(self, &user)
  }
}

func collectEventIds(list []EventRecord) []int {
  event_ids := make([]int, len(list))
  for i, item := range list {
    event_ids[i] = item.Id
  }
  return event_ids
}

func eventClass(team TeamRecord, count int, status int) string {
  switch {
  case status == eventStatusCanceled:
    return "skip"
  case count < team.UsersMin:
    return "under"
  case count >= team.UsersMax && team.UsersMax > 0:
    return "over"
  default:
    return "fits"
  }
}
