package main

import (
  "time"
)

type Locale struct {
  Date string
}

const (
  serverName = "Nykredit Fitness"
  serverHost = "nk-fitness.dk"
  serverRoot = "http://nk-fitness.dk"
  serverPort = ":8001"
  adminEmail = "fitness.glaskuben@nykredit.dk" 

  cookieHost = ""
  cookieAuth = "nk-fitness$$" // 32 bytes
  cookieEncr = "nk-fitness$$" // 32 bytes
  cookieLife = 1 * time.Hour

  emailsRoot = "https://api.mailgun.net/v3/nk-fitness.dk/messages"
  emailsUser = "api"
  emailsPass = ""
  emailsFrom = "info@nk-fitness.dk"

  smsHost    = "http://sms.coolsmsc.dk:8080/sendsms.php"
  smsUser    = "sms"
  smsPass    = ""
  smsFrom    = "NK Fitness"

  cancelCheck = "0 5 * * * *"
  cancelAhead = 5 * time.Hour // in 5-6 hours
  autoConfirm = true
  gracePeriod = 2 * time.Hour // if not autoConfirm
  smsInPeriod = 24 * time.Hour

  siteSecret  = "nk-fitness$$" // 32 bytes
  sessionKey  = "session"
  bcryptCost  = 10
  minPassLen  = 6
  expireLink  = 2 * time.Hour
  defaultLang = "da"
  defaultPage = "/calendar/week" // not "/"
  defaultDate = "2015-01-01"
)

var locales = map[string]Locale{
  "da": Locale{Date: "02/01 2006"},
  "en": Locale{Date: "2006-01-02"},
}

const (
  panicError  = "Critical error happened, please contact website admin"
  permitError = "You are not authorized to view this page"

  timeFormat  = "15:04"
  dateFormat  = "2006-01-02" // db
  fullFormat  = "2006-01-02 15:04:05" // db
)
