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

// --- controller helpers ---

func setPage(c *gin.Context) {
  p, _ := c.Get("page")
  if page, ok := p.(string); ok {
    return
  }
  path := c.Request.URL.Path
  idx := strings.LastIndex(path, "/")
  c.Set("page", path[idx+1:])
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


// --- session methods ---

func setSessionAuthInfo(c *gin.Context, user_id int) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.Values["user_id"] = user_id
}

// TODO: moreinfo, struct?
func getSessionAuthInfo(c *gin.Context) *int {
  session, _ := cookie.Get(c.Request, sessionKey)
  log.Printf("=> SESSION\n   %#v, %#v\n", session.Values, session.Options) // <<< DEBUG
  if user_id, ok := session.Values["user_id"].(int); ok {
    return &user_id
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

func currentUserId(c *gin.Context) (int, bool) {
  user, _ := c.Get("self")
  user_id, ok := user.(int)
  if ok {
    return user_id, true
  } else {
    log.Printf("[APP] CUR_USER error: user=%#v\n", user)
    return -1, false
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
