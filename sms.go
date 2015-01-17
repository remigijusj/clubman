package main

import (
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
)

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
