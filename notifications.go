package main

import (
  "bytes"
  "fmt"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"

  "gopkg.in/gomail.v1"
)

var mails *template.Template

func loadMailTemplates(pattern string) {
  mails = template.Must(template.New("").Funcs(helpers).ParseGlob("mails/*"))
}

func sendEmail(to, subject, body string) error {
  msg := gomail.NewMessage()
  msg.SetHeader("From", emailsFrom)
  msg.SetHeader("To", to)
  msg.SetHeader("Subject", subject)
  msg.SetBody("text/html", body)

  mailer := gomail.NewMailer(emailsHost, emailsUser, emailsPass, emailsPort)
  return mailer.Send(msg)
}

func sendSMS(mobile, message string) {
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
    return
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Printf("[APP] SMS error: response %s\n", err)
    return
  }
  val, err := url.ParseQuery(string(body))
  if err != nil {
    log.Printf("[APP] SMS error: response body %s (%s)\n", err, body)
    return
  }
  status := val.Get("status")
  if status != "success" {
    log.Printf("[APP] SMS error: response status %s\n", status)
  }

  return
}

// NOTE: delayed
func sendResetLinkEmail(lang, email, token string) {
  obj := map[string]string{
    "host": serverHost,
    "url":  fmt.Sprintf("%s/resets?email=%s&token=%s", serverRoot, url.QueryEscape(email), token),
  }
  var buf bytes.Buffer
  mails.Lookup("password_reset").Funcs(transHelpers[lang]).Execute(&buf, obj)
  message := buf.String()

  subject := T(lang, "Password reset for %s", serverHost)

  err := sendEmail(email, subject, message)
  if err != nil {
    log.Printf("[APP] SEND-RESET error: %s, token=%s, email=%s\n", err, token, email)
  }
}

func sendEventConfirmLink(user_id, event_id int) {
  user, err := userContact(user_id)
  if err != nil {
    log.Printf("[APP] SEND-EVENT-CONFIRM: %v, %d, %d, %v\n", err, user_id, event_id, user)
    return
  }
  event, err := fetchEventInfo(event_id)
  if err != nil {
    log.Printf("[APP] SEND-EVENT-CONFIRM: %v, %d, %d, %v\n", err, user_id, event_id, event)
    return
  }
  sendEventConfirmLinkEmail(user.Language, user.Email, event_id, &event)
}

func sendEventConfirmLinkEmail(lang, email string, event_id int, event *EventInfo) {
  obj := map[string]interface{}{
    "event": event,
    "url":  fmt.Sprintf("%s/assignments/confirm/%d", serverRoot, event_id),
    "lang": lang,
  }
  var buf bytes.Buffer
  mails.Lookup("event_confirm").Funcs(transHelpers[lang]).Execute(&buf, obj)
  message := buf.String()

  subject := T(lang, "Confirm subscription for %s", event.Name)

  err := sendEmail(email, subject, message)
  if err != nil {
    log.Printf("[APP] SEND-EVENT-CONFIRM-EMAIL error: %s\n", err)
  }
}
