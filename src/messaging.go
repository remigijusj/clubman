package main

import (
  "bytes"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "strings"
  "time"
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
  return diff >= 0 && diff < conf.SmsInPeriod.Duration
}

var mails *template.Template

func loadMailTemplates(pattern string) {
  mails = template.Must(template.New("").Funcs(helpers).ParseGlob("mails/*"))
}

func sendEmail(to, subject, body string, args ...string) bool {
  if debugMode() {
    log.Printf("[APP] DEBUG EMAIL to %s: %s\n", to, subject)
    return true
  }

  data := url.Values{}
  data.Add("from",    conf.EmailsFrom)
  data.Add("to",      to)
  data.Add("subject", subject)
  data.Add("html",    body)
  if len(args) > 0 && args[0] != "" {
    data.Add("h:Reply-To", args[0])
  }

  req, err := http.NewRequest("POST", conf.EmailsRoot, strings.NewReader(data.Encode()))
  if err != nil {
    log.Printf("[APP] EMAIL request error: %v, %s, %s\n", err, to, subject)
  }
  req.SetBasicAuth(conf.EmailsUser, conf.EmailsPass)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  client := &http.Client{}
  resp, err := client.Do(req)
  _ = resp

  if err != nil {
    log.Printf("[APP] EMAIL sending error: %v, %s, %s\n", err, to, subject)
  }
  return err == nil
/*
  defer resp.Body.Close()
  b, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Printf("[APP] EMAIL error: %v, %s, %s, [%s]\n", err, to, subject, b)
  }
*/
}

func sendSMS(mobile, message string) bool {
  if debugMode() {
    log.Printf("[APP] DEBUG SMS to %s: %s\n", mobile, message)
    return true
  }

  v := url.Values{}

  v.Set("username",   conf.SmsUser)
  v.Set("password",   conf.SmsPass)
  v.Set("from",       conf.SmsFrom)
  v.Set("to",         mobile)
  v.Set("message",    message)
  v.Set("charset",    "utf-8")
  v.Set("resulttype", "urlencoded")

  uri := conf.SmsHost + "?" + v.Encode()

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
    log.Printf("[APP] SMS error: response status %s (%s)\n", status, body)
  }

  return true
}

func compileMessage(name, lang string, data interface{}) string {
  var buf bytes.Buffer
  mails.Lookup(name).Funcs(transHelpers[lang]).Execute(&buf, data)
  return string(bytes.TrimSpace(buf.Bytes()))
}
