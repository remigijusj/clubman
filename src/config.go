package main

import (
  "io/ioutil"
  "fmt"
  "log"
  "os"
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
  AdminEmail  string
  LogoImgUrl  string

  CookieHost  string
  CookieAuth  string
  CookieEncr  string
  CookieLife  duration

  EmailsRoot  string
  EmailsKey   string
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
  data, err := readConfig()
  if err != nil {
    panic(err)
  }
  err = toml.Unmarshal(data, &conf)
  if err != nil {
    panic(err)
  }
  validateConfig()
}

func readConfig() ([]byte, error) {
  if _, err := os.Stat(configFile); os.IsNotExist(err) {
    value := os.Getenv(configVar)
    if value == "" {
      return nil, fmt.Errorf("Both %s and env %s are missing", configFile, configVar)
    } else {
      return []byte(value), nil
    }
  } else {
    return ioutil.ReadFile(configFile)
  }
}

func validateConfig() {
  // TODO: implement
  if debugMode() {
    log.Printf("=> CONFIG: %#v\n\n", *conf)
  }
}
