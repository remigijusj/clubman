package main

import (
  "bytes"
  "fmt"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "time"

  "gopkg.in/gomail.v1"
)

const (
  contactEmail = 1
  contactSMS   = 2
)

var mails *template.Template

func (self UserContact) chooseMethod() int {
  now := time.Now().Format(timeFormat)
  if now < workdayFrom || now >= workdayTill {
    return contactSMS
  } else {
    return contactEmail
  }
}

func loadMailTemplates(pattern string) {
  mails = template.Must(template.New("").Funcs(helpers).ParseGlob("mails/*"))
}

func sendEmail(to, subject, body string) bool {
  msg := gomail.NewMessage()
  msg.SetHeader("From", emailsFrom)
  msg.SetHeader("To", to)
  msg.SetHeader("Subject", subject)
  msg.SetBody("text/html", body)

  mailer := gomail.NewMailer(emailsHost, emailsUser, emailsPass, emailsPort)
  err := mailer.Send(msg)
  if err != nil {
    log.Printf("[APP] EMAIL error: %v, %s, %s\n", err, to, subject)
  }
  return err == nil
}

func sendSMS(mobile, message string) bool {
  v := url.Values{}

  v.Set("username",   smsUser)
  v.Set("password",   smsPass)
  v.Set("from",       smsFrom)
  v.Set("to",         mobile)
  v.Set("message",    message)
  v.Set("charset",    "utf-8")
  v.Set("resulttype", "urlencoded")

  uri := smsHost + "?" + v.Encode()

  resp, err := http.Get(uri)
  if err != nil {
    log.Printf("[APP] SMS error: sending %s\n", err)
    return false
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Printf("[APP] SMS error: response %s\n", err)
    return false
  }
  val, err := url.ParseQuery(string(body))
  if err != nil {
    log.Printf("[APP] SMS error: response body %s (%s)\n", err, body)
    return false
  }
  status := val.Get("status")
  if status != "success" {
    log.Printf("[APP] SMS error: response status %s\n", status)
  }

  return true
}

func compileMessage(name, lang string, data interface{}) string {
  var buf bytes.Buffer
  mails.Lookup(name).Funcs(transHelpers[lang]).Execute(&buf, data)
  return string(bytes.TrimSpace(buf.Bytes()))
}

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

func notifyEventConfirm(event *EventInfo, user *UserContact) {
  switch user.chooseMethod() {
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

func notifyEventCancel(event *EventInfo, users []UserContact) {
  for _, user := range users {
    notifyEventUserCancel(event, &user)
  }
}

func notifyEventUserCancel(event *EventInfo, user *UserContact) {
  switch user.chooseMethod() {
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

func notifyEventMultiCancel(data *TeamEventsData, team *TeamForm, users []UserContact) {
  for _, user := range users {
    notifyEventUserMultiCancel(data, team, &user)
  }
}

func notifyEventUserMultiCancel(data *TeamEventsData, team *TeamForm, user *UserContact) {
  sendEventMultiCancelEmail(user.Email, user.Language, data, team)
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

