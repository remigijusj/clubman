package main

import (
  "log"
  "github.com/gin-gonic/gin"
)

// --- controller helpers ---

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
