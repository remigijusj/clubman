package main

import (
  "bytes"
  "html/template"
  "log"
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

// TODO: localize body
func sendResetEmail(lang, email, token string) {
  obj := map[string]string{
    "host": serverHost,
    "url":  serverRoot+"/resets?email="+url.QueryEscape(email)+"&token="+token,
  }
  var buf bytes.Buffer
  mails.Lookup("password_reset").Execute(&buf, obj)
  message := buf.String()

  subject := T(lang, "Password reset for %{host}", serverHost)

  err := sendEmail(email, subject, message)
  if err != nil {
    log.Printf("[APP] RESET-EMAIL error: %s, token=%s, email=%s\n", err, token, email)
  }
}
