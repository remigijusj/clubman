package main

import (
  "bytes"
  "html/template"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "net/url"
  "time"
)

const (
  contactEmail = 1
  contactSMS   = 2
)

type EmailAddress struct {
  Email   string  `json:"email"`
}

type Personalization struct {
  To      []EmailAddress  `json:"to"`
}

type EmailContent struct {
  Type    string  `json:"type"`
  Value   string  `json:"value"`
}

type EmailPayload struct {
  Personalizations []Personalization `json:"personalizations"`
  From    EmailAddress   `json:"from"`
  ReplyTo EmailAddress   `json:"reply_to,omitempty"`
  Subject string         `json:"subject"`
  Content []EmailContent `json:"content"`
}

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
    logPrintf("DEBUG EMAIL to %s: %s\n", to, subject)
    return true
  }

  content := EmailContent{
    Type: "text/html",
    Value: body,
  }
  email_to := EmailAddress{
    Email: to,
  }
  personalization := Personalization{
    To:    []EmailAddress{email_to},
  }
  emailPayload := EmailPayload{
    Personalizations: []Personalization{personalization},
    From:    EmailAddress{Email: conf.EmailsFrom},
    ReplyTo: EmailAddress{Email: conf.EmailsFrom},
    Subject: subject,
    Content: []EmailContent{content},
  }
  if len(args) > 0 && args[0] != "" {
    emailPayload.ReplyTo.Email = args[0]
  }
  var jsonPayload []byte
  jsonPayload, err := json.Marshal(emailPayload)
  if err != nil {
    logPrintf("EMAIL build error: %v, %s, %s, %s\n", err, to, subject, string(jsonPayload))
  }
  logPrintf("SENDGRID PAYLOAD: %s\n", string(jsonPayload))

  req, err := http.NewRequest("POST", conf.EmailsRoot, bytes.NewReader(jsonPayload))
  if err != nil {
    logPrintf("EMAIL request error: %v, %s\n", err, string(jsonPayload))
  }
  req.Header.Set("Authorization", "Bearer " + conf.EmailsKey)
  req.Header.Set("Content-Type", "application/json")
  client := &http.Client{}
  resp, err := client.Do(req)
  _ = resp

  if err != nil {
    logPrintf("EMAIL sending error: %v, %s, %s\n", err, to, subject)
  }
  return err == nil
/*
  defer resp.Body.Close()
  b, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    logPrintf("EMAIL error: %v, %s, %s, [%s]\n", err, to, subject, b)
  }
*/
}

func sendSMS(mobile, message string) bool {
  if debugMode() {
    logPrintf("DEBUG SMS to %s: %s\n", mobile, message)
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
    logPrintf("SMS error: sending %s\n", err)
    return false
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    logPrintf("SMS error: response %s\n", err)
    return false
  }
  val, err := url.ParseQuery(string(body))
  if err != nil {
    logPrintf("SMS error: response body %s (%s)\n", err, body)
    return false
  }
  status := val.Get("status")
  if status != "success" {
    logPrintf("SMS error: response status %s (%s)\n", status, body)
  }

  return true
}

func compileMessage(name, lang string, data interface{}) string {
  var buf bytes.Buffer
  mails.Lookup(name).Funcs(transHelpers[lang]).Execute(&buf, data)
  return string(bytes.TrimSpace(buf.Bytes()))
}
