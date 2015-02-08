package main

import (
  "fmt"
  "net/url"
)

// NOTE: delayed
func sendResetLinkEmail(email, lang, token string) {
  subject := T(lang, "Password reset for %s", serverHost)

  obj := map[string]string{
    "host": serverHost,
    "url":  fmt.Sprintf("%s/resets?email=%s&token=%s", serverRoot, url.QueryEscape(email), token),
  }
  message := compileMessage("password_reset_email", lang, obj)

  sendEmail(email, subject, message)
}

func notifyEventToConfirm(event *EventInfo, user *UserContact) {
  switch user.chooseMethod(event.StartAt) {
  case contactEmail:
    sendEventConfirmLinkEmail(user.Email, user.Language, event)
  case contactSMS:
    sendEventConfirmLinkSMS(user.Mobile, user.Language, event)
  }
}

func sendEventConfirmLinkEmail(email, lang string, event *EventInfo) {
  subject := T(lang, "Confirm subscription for %s", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
    "url":  fmt.Sprintf("%s/assignments/confirm/%d", serverRoot, event.Id),
  }
  message := compileMessage("event_confirm_email", lang, obj)

  sendEmail(email, subject, message)
}

func sendEventConfirmLinkSMS(mobile, lang string, event *EventInfo) {
  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
    "host": serverHost,
  }
  message := compileMessage("event_confirm_sms", lang, obj)

  sendSMS(mobile, message)
}

func notifyEventConfirmed(event *EventInfo, user *UserContact) {
  switch user.chooseMethod(event.StartAt) {
  case contactEmail:
    sendEventConfirmedEmail(user.Email, user.Language, event)
  case contactSMS:
    sendEventConfirmedSMS(user.Mobile, user.Language, event)
  }
}

func sendEventConfirmedEmail(email, lang string, event *EventInfo) {
  subject := T(lang, "Subscription for %s confirmed", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
    "url":  fmt.Sprintf("%s/assignments/delete/%d", serverRoot, event.Id),
  }
  message := compileMessage("event_confirmed_email", lang, obj)

  sendEmail(email, subject, message)
}

func sendEventConfirmedSMS(mobile, lang string, event *EventInfo) {
  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
    "host": serverHost,
  }
  message := compileMessage("event_confirmed_sms", lang, obj)

  sendSMS(mobile, message)
}

func notifyEventWaitingUp(event *EventInfo, user *UserContact, num int) {
  sendEventWaitingUp(user.Email, user.Language, event, num)
}

func sendEventWaitingUp(email, lang string, event *EventInfo, num int) {
  subject := T(lang, "Up in waiting list for %s", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
    "num":  num,
    "url":  fmt.Sprintf("%s/assignments/delete/%d", serverRoot, event.Id),
  }
  message := compileMessage("event_waiting_up_email", lang, obj)

  sendEmail(email, subject, message)
}

func notifyEventCancel(event *EventInfo, user *UserContact) {
  switch user.chooseMethod(event.StartAt) {
  case contactEmail:
    sendEventCancelEmail(user.Email, user.Language, event)
  case contactSMS:
    sendEventCancelSMS(user.Mobile, user.Language, event)
  }
}

func sendEventCancelEmail(email, lang string, event *EventInfo) {
  subject := T(lang, "%s is canceled", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
  }
  message := compileMessage("event_cancel_email", lang, obj)

  sendEmail(email, subject, message)
}

func sendEventCancelSMS(mobile, lang string, event *EventInfo) {
  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
  }
  message := compileMessage("event_cancel_sms", lang, obj)

  sendSMS(mobile, message)
}

func notifyEventUpdate(event *EventInfo, user *UserContact) {
  switch user.chooseMethod(event.StartAt) {
  case contactEmail:
    sendEventUpdateEmail(user.Email, user.Language, event)
  case contactSMS:
    sendEventUpdateSMS(user.Mobile, user.Language, event)
  }
}

func sendEventUpdateEmail(email, lang string, event *EventInfo) {
  subject := T(lang, "%s is updated", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
  }
  message := compileMessage("event_update_email", lang, obj)

  sendEmail(email, subject, message)
}

func sendEventUpdateSMS(mobile, lang string, event *EventInfo) {
  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
  }
  message := compileMessage("event_update_sms", lang, obj)

  sendSMS(mobile, message)
}

func notifyEventMultiUpdate(data *TeamEventsData, user *UserContact, team *TeamForm, near bool) {
  sendEventMultiUpdateEmail(user.Email, user.Language, data, team)
  if near {
    sendEventMultiUpdateSMS(user.Email, user.Language, data, team)
  }
}

func sendEventMultiUpdateEmail(email, lang string, data *TeamEventsData, team *TeamForm) {
  subject := T(lang, "Multiple %s events updated", team.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "data": *data,
    "team": *team,
  }
  message := compileMessage("event_update_multi_email", lang, obj)

  sendEmail(email, subject, message)
}

func notifyEventMultiCancel(data *TeamEventsData, user *UserContact, team *TeamForm, near bool) {
  sendEventMultiCancelEmail(user.Email, user.Language, data, team)
  if near {
    sendEventMultiCancelSMS(user.Email, user.Language, data, team)
  }
}

func sendEventMultiUpdateSMS(mobile, lang string, data *TeamEventsData, team *TeamForm) {
  obj := map[string]interface{}{
    "lang": lang,
    "data": *data,
    "team": *team,
  }
  message := compileMessage("event_update_multi_sms", lang, obj)

  sendSMS(mobile, message)
}

func sendEventMultiCancelEmail(email, lang string, data *TeamEventsData, team *TeamForm) {
  subject := T(lang, "Multiple %s events canceled", team.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "data": *data,
    "team": *team,
  }
  message := compileMessage("event_cancel_multi_email", lang, obj)

  sendEmail(email, subject, message)
}

func sendEventMultiCancelSMS(mobile, lang string, data *TeamEventsData, team *TeamForm) {
  obj := map[string]interface{}{
    "lang": lang,
    "data": *data,
    "team": *team,
  }
  message := compileMessage("event_cancel_multi_sms", lang, obj)

  sendSMS(mobile, message)
}

func sendAssignmentCreatedEmail(email, lang string, event *EventInfo, confirmed bool) {
  subject := T(lang, "Subscribed for %s", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
    "confirmed": confirmed,
  }
  message := compileMessage("assignment_create_email", lang, obj)

  sendEmail(email, subject, message)
}

func sendAssignmentDeletedEmail(email, lang string, event *EventInfo) {
  subject := T(lang, "Canceled from %s", event.Name)
  subject = fmt.Sprintf("[%s] %s", serverName, subject)

  obj := map[string]interface{}{
    "lang": lang,
    "event": event,
  }
  message := compileMessage("assignment_delete_email", lang, obj)

  sendEmail(email, subject, message)
}
