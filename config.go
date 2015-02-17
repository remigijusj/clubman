package main

import (
  "time"
)

const (
  serverName = "Nykredit Fitness"
  serverHost = "demo.nk-fitness.dk"
  serverRoot = "http://demo.nk-fitness.dk"
  serverPort = ":8001"

  cookieHost = ""
  cookieAuth = "nk-fitness#nk-fitness#nk-fitness" // 32 bytes
  cookieEncr = "nk-fitness$nk-fitness$nk-fitness" // 32 bytes
  cookieLife = 3600 * 1 // 1 hours

  emailsHost = "smtp.gmail.com"
  emailsUser = ""
  emailsPass = ""
  emailsPort = 587
  emailsFrom = ""

  smsHost    = "http://sms.coolsmsc.dk:8080/sendsms.php"
  smsUser    = ""
  smsPass    = ""
  smsFrom    = ""

  cancelCheck = "0 5 * * * *"
  cancelAhead = 5 * time.Hour // in 5-6 hours
  autoConfirm = true
  gracePeriod = 2 * time.Hour // if not autoConfirm
  smsInPeriod = 24 * time.Hour

  siteSecret  = ""
  sessionKey  = "session"
  bcryptCost  = 10
  minPassLen  = 6
  expireLink  = 2 * time.Hour
  defaultLang = "da"
  defaultPage = "/calendar/week" // not "/"
  defaultDate = "2015-01-01"
  timeFormat  = "15:04"
  dateFormat  = "2006-01-02" // db
  fullFormat  = "2006-01-02 15:04:05" // db
  panicError  = "Critical error happened, please contact website admin"
  permitError = "You are not authorized to view this page"

  reloadTmpl  = true // DEBUG mode
)

var languages = []string{"da", "en"}

var dateFormats = map[string]string{
  "da": "02/01 2006",
  "en": "2006-01-02",
}
