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

func displayError(c *gin.Context, message string) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"warning", message})
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

// --- authorization helpers ---

func currentUserId(c *gin.Context) (int, bool) {
  self, _ := c.Get("self")
  auth, ok := self.(AuthInfo)
  if ok {
    return auth.Id, true
  } else {
    log.Printf("[APP] CUR_USER error: self=%#v\n", self)
    return -1, false
  }
}

func isAuthenticated(c *gin.Context) bool {
  _, err := c.Get("self")
  return err == nil
}

func isAdmin(c *gin.Context) bool {
  self, _ := c.Get("self")
  auth, ok := self.(AuthInfo)
  log.Printf("=> ISADMIN: %#v, %v\n", auth, ok)
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
