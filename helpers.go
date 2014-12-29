package main

import (
  "crypto/rand"
  "fmt"
  "log"
  "strings"

  "code.google.com/p/go.crypto/bcrypt"
  "github.com/gin-gonic/gin"
)

const (
  sessionKey = "session"
  bcryptCost = 10
)

type Alert struct {
  Kind, Message string
}

type AuthInfo struct {
  Id     int
  Name   string
  Status int
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

func showError(c *gin.Context, err error, args ...interface{}) {
  message := err.Error()
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"warning", message})
  }
  if len(args) > 0 {
    setFlashedForm(c, args[0])
  }
  c.Redirect(302, c.Request.URL.Path)
}

func forwardTo(c *gin.Context, route string, message string) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"success", message})
  }
  c.Redirect(302, route)
}

func forwardWarning(c *gin.Context, route string, message string) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"warning", message})
  }
  c.Redirect(302, route)
}


// --- session methods ---

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

func deleteSession(c *gin.Context) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.Options.MaxAge = -1
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

// --- authorization helpers ---

func currentUser(c *gin.Context) *AuthInfo {
  self, _ := c.Get("self")
  auth, ok := self.(AuthInfo)
  if ok {
    return &auth
  } else {
    log.Printf("[APP] CUR_USER error: self=%#v\n", self)
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
    return auth.Status == StatusAdmin
  } else {
    return false
  }
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
