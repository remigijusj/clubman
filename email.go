package main

import (
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

// TODO: use external template
func sendResetEmail(email, token string) {
  url := serverRoot+"/resets?email="+url.QueryEscape(email)+"&token="+token

  subject := "Password reset for "+serverHost
  message := "<p><b>You have requested password reset for "+serverHost+"</b></p>"
  message += "<p>Please click the link and change your password:</p>"
  message += `<p><a href="`+url+`">`+url+`</a></p>`

  err := sendEmail(email, subject, message)
  if err != nil {
    log.Printf("[APP] RESET-EMAIL error: %s, token=%s, email=%s\n", err, token, email)
  }
}
