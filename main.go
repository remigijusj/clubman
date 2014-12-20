package main

import (
  "database/sql"
  "encoding/gob"
  "fmt"
  //"html/template"
  //"os"

  "github.com/gin-gonic/gin"
  "github.com/gorilla/sessions"
  _ "github.com/mattn/go-sqlite3" // tdm-gcc
)

const (
  cookieHost = ""
  cookieAuth = "nk-fitness#nk-fitness#nk-fitness" // 32 bytes
  cookieEncr = "nk-fitness$nk-fitness$nk-fitness" // 32 bytes
  sessionKey = "session"
)

var db *sql.DB
var query map[string]*sql.Stmt
var cookie *sessions.CookieStore

type Alert struct {
  Kind, Message string
}

func init() {
  db, _ = sql.Open("sqlite3", "./main.db")
  prepareSQL()
  gob.Register(&Alert{})
  cookie = sessions.NewCookieStore([]byte(cookieAuth), []byte(cookieEncr))
  cookie.Options = &sessions.Options{
    Domain:   cookieHost,
    Path:     "/",
    MaxAge:   3600 * 8, // 8 hours
    HttpOnly: false,
    Secure:   false,
  }
}

func prepareSQL() {
  query = make(map[string]*sql.Stmt, 5)
  query["user_password"], _ = db.Prepare("SELECT id, password FROM users WHERE email=?")
  query["user_password_reset"], _ = db.Prepare("UPDATE users SET password=? WHERE email=?")
  query["user_insert"], _ = db.Prepare("INSERT INTO users(name, email, mobile, password) values (?, ?, ?, ?)")
  query["user_update"], _ = db.Prepare("UPDATE users SET name=?, email=?, mobile=?, password=?, language=? WHERE email=?")
}

func main() {
  r := gin.Default()
  r.LoadHTMLFiles("main.tmpl")

  r.GET("/", displayPage) // redirect to welcome?
  r.Static("/css", "./css")
  r.Static("/img", "./img")
  r.Static("/js", "./js")

  r.GET("/login", displayPage)
  r.GET("/forgot", displayPage)
  // r.GET("/reset_password", displayPage)
  r.GET("/signup", displayPage)
  r.GET("/profile", displayPage)
  r.GET("/logout", handleLogout)

  r.GET("/list", displayPage)
  r.GET("/calendar", displayPage)

  r.POST("/login", handleLogin)
  r.POST("/forgot", handleForgot)
  r.POST("/signup", handleSignup)
  r.POST("/profile", handleProfile)

  r.Run(":8001")
  // <<< query..Close(); db.Close()
}

func displayPage(c *gin.Context) {
  page := c.Request.URL.Path[1:]
  user := getSessionAuthInfo(c)
  if ok := authorizeView(user, page); !ok {
    setSessionAlert(c, &Alert{"warning", "You are not authorized to view this page."})
    c.Redirect(302, "/login")
    return
  }
  alert := getSessionAlert(c)
  obj := gin.H{"page": page, "user": user, "alert": alert}
  c.HTML(200, "main.tmpl", obj)
}

// --- helper methods ---

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

func authorizeView(user *int, page string) bool {
  // <<< TODO
  if user == nil && page == "profile" {
    return false
  }
  return true
}

// --- session methods ---

func setSessionAuthInfo(c *gin.Context, user_id int) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.Values["user_id"] = user_id
}

func getSessionAuthInfo(c *gin.Context) *int {
  session, _ := cookie.Get(c.Request, sessionKey)
  fmt.Printf("SESSION: %#v, %#v\n", session.Values, session.Options) // <<< DEBUG
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

// --- TODO list ---
//   display form values (profile, redirected)
//   authorization: inner pages, post handlers
//   check passwords: bcypt, scrypt, salt?
//   remind email - reset password
//   form validation in JS
//   user status handling
//   pjax double render
//   proper logging
