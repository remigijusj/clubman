package main

import (
  "bytes"
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

// NOTE: user not used now; might have preferred method attr
func (self UserContact) chooseMethod(when time.Time) int {
  if isNear(when) {
    return contactSMS
  } else {
    return contactEmail
  }
}

func isNear(when time.Time) bool {
  diff := when.Sub(time.Now())
  return diff >= 0 && diff < smsInPeriod
}

var mails *template.Template

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
