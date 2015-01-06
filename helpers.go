package main

import (
  "crypto/rand"
  "fmt"
  "log"
  "strings"

  "code.google.com/p/go.crypto/bcrypt"
  "github.com/gin-gonic/gin"
)

type Alert struct {
  Kind, Message string
}

type AuthInfo struct {
  Id       int
  Name     string
  Status   int
  Language string
}

type ErrorWithArgs struct {
  Message  string
  Args     []interface{}
}

type SimpleRecord struct {
  Id       int
  Text     string
}

// --- controller helpers ---

func setPage(c *gin.Context) {
  if _, err := c.Get("page"); err == nil {
    return
  }
  tokens := strings.Split(c.Request.URL.Path[1:], "/")
  switch {
  case len(tokens) == 1:
    c.Set("page", tokens[0])
  case len(tokens) > 1:
    c.Set("page", tokens[0] + "_" + tokens[1])
  }
}

// NOTE: extra is used for optional form reference
func showError(c *gin.Context, err error, extra ...interface{}) {
  message := err.Error()
  if len(message) > 0 {
    var args []interface{}
    if erra, ok := err.(*ErrorWithArgs); ok {
      args = erra.Args
    }
    setSessionAlert(c, &Alert{"warning", TC(c, message, args...)})
  }
  if len(extra) > 0 && extra[0] != nil {
    setFlashedForm(c, extra[0])
  }
  var route string
  if len(extra) > 1 {
    route = extra[1].(string)
  } else {
    route = c.Request.URL.Path
  }
  c.Redirect(302, route)
}

func forwardTo(c *gin.Context, route string, message string, args ...interface{}) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"success", TC(c, message, args...)})
  }
  c.Redirect(302, route)
}

func forwardWarning(c *gin.Context, route string, message string, args ...interface{}) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"warning", TC(c, message, args...)})
  }
  c.Redirect(302, route)
}

// --- session helpers ---

func setSessionAuthInfo(c *gin.Context, auth *AuthInfo) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.Values["auth"] = auth
}

func getSessionAuthInfo(c *gin.Context) *AuthInfo {
  session, _ := cookie.Get(c.Request, sessionKey)
  log.Printf("=> SESSION\n   %#v, %#v\n", session.Values, session.Options) // <<< DEBUG
  if auth, ok := session.Values["auth"].(*AuthInfo); ok {
    return auth
  }
  return nil
}

func setSessionAlert(c *gin.Context, alert *Alert) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.AddFlash(alert)
}

func getSessionAlert(c *gin.Context) *Alert {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  if flashes := session.Flashes(); len(flashes) > 0 {
    if flash, ok := flashes[0].(*Alert); ok {
      return flash
    }
  }
  return nil
}

func setFlashedForm(c *gin.Context, form interface{}) {
  session, _ := cookie.Get(c.Request, "flash-form")
  defer session.Save(c.Request, c.Writer)
  session.Values["form"] = form
}

func getFlashedForm(c *gin.Context) interface{} {
  session, _ := cookie.Get(c.Request, "flash-form")
  defer session.Save(c.Request, c.Writer)
  session.Options.MaxAge = -1
  return session.Values["form"]
}

func deleteSession(c *gin.Context) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.Options.MaxAge = -1
}

// --- authorization helpers ---

func currentUser(c *gin.Context) *AuthInfo {
  self, _ := c.Get("self")
  auth, ok := self.(AuthInfo)
  if ok {
    return &auth
  } else {
    return nil
  }
}

func isAuthenticated(c *gin.Context) bool {
  _, err := c.Get("self")
  return err == nil
}

func isAdmin(c *gin.Context) bool {
  self, _ := c.Get("self")
  auth, ok := self.(AuthInfo)
  if ok {
    return auth.Status == userStatusAdmin
  } else {
    return false
  }
}

func (self AuthInfo) IsAdmin() bool {
  return self.Status == userStatusAdmin
}

// --- password-related ---

func hashPassword(password string) string {
  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
  if err != nil {
    log.Printf("[APP] BCRYPT error: %s\n", err)
  }
  return string(hash)
}

func comparePassword(stored, given string) bool {
  err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(given))
  if err != nil {
    log.Printf("[APP] BCRYPT error: %s\n", err)
  }
  return err == nil
}

func generateToken(size int) string {
  b := make([]byte, size)
  rand.Read(b)
  return fmt.Sprintf("%x", b)
}

// --- error helpers ---

// NOTE: fmt.Sprintf no fit because we use reuslt for i18n
func (err *ErrorWithArgs) Error() string {
  return err.Message
}

func errorWithA(message string, args ...interface{}) *ErrorWithArgs {
  return &ErrorWithArgs{message, args}
}

// --- miscelaneous ---

func listRecords(name string, args ...interface{}) []SimpleRecord {
  list := []SimpleRecord{}
  rows, err := query[name].Query(args...)
  if err != nil {
    log.Printf("[APP] USER-LIST-STATUS error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item SimpleRecord
    err := rows.Scan(&item.Id, &item.Text)
    if err != nil {
      log.Printf("[APP] USER-LIST-STATUS error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-LIST error: %s\n", err)
  }
  return list
}
