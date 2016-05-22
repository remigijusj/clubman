package main

import (
  "io/ioutil"
  "log"
  "time"

  "github.com/BurntSushi/toml"
)

type duration struct {
  time.Duration
}

type Locale struct {
  Date string
}

type Conf struct {
  ServerName  string
  ServerHost  string
  ServerRoot  string
  ServerPort  string
  AdminEmail  string
  LogoImgUrl  string

  CookieHost  string
  CookieAuth  string
  CookieEncr  string
  CookieLife  duration

  EmailsRoot  string
  EmailsUser  string
  EmailsPass  string
  EmailsFrom  string

  SmsHost     string
  SmsUser     string
  SmsPass     string
  SmsFrom     string

  CancelCheck string
  CancelAhead duration
  CancelRange duration
  AutoConfirm bool
  GracePeriod duration
  SmsInPeriod duration

  SiteSecret  string
  SessionKey  string
  BcryptCost  int
  MinPassLen  int
  ExpireLink  duration

  DefaultPage string
  DefaultLang string
  DefaultDate string

  Locales map[string]Locale
}

func (d *duration) UnmarshalText(text []byte) error {
  var err error
  d.Duration, err = time.ParseDuration(string(text))
  return err
}

func prepareConfig() {
  var err error
  data, err := ioutil.ReadFile("config.toml")
  if err != nil {
    panic(err)
  }
  err = toml.Unmarshal(data, &conf)
  if err != nil {
    panic(err)
  }
  validateConfig()
}

func validateConfig() {
  // TODO: implement
  if debugMode() {
    log.Printf("=> CONFIG: %#v\n\n", *conf)
  }
}
