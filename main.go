package main

import (
  "database/sql"
  "encoding/gob"
  "log"
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
  log.SetFlags(0)
  gob.Register(&Alert{})

  db, _ = sql.Open("sqlite3", "./main.db")
  prepareSQL()

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
  query["password_select"], _ = db.Prepare("SELECT id, password FROM users WHERE email=?")
  query["password_update"], _ = db.Prepare("UPDATE users SET password=? WHERE email=?")
  query["user_select"], _ = db.Prepare("SELECT name, email, mobile, language FROM users WHERE id=?")
  query["user_insert"], _ = db.Prepare("INSERT INTO users(name, email, mobile, password) values (?, ?, ?, ?)")
  query["user_update"], _ = db.Prepare("UPDATE users SET name=?, email=?, mobile=?, password=?, language=? WHERE email=?")
}

func main() {
  r := gin.Default()
  r.LoadHTMLFiles("main.tmpl")

  // TODO: combine to 1 public dir
  s := r.Group("/")
  {
    s.Handlers = s.Handlers[:1] // removing Logger from Default
    s.Static("/css", "./css")
    s.Static("/img", "./img")
    s.Static("/js", "./js")
  }

  r.GET("/login", displayPage)
  r.GET("/forgot", displayPage)
  // r.GET("/reset_password", displayPage)
  r.GET("/signup", displayPage)
  r.POST("/login", handleLogin)
  r.POST("/forgot", handleForgot)
  r.POST("/signup", handleSignup)

  a := r.Group("/", authRequired())
  {
    // TODO: redirect to welcome?
    a.GET("/", displayPage)
    a.GET("/logout", handleLogout)
    a.GET("/profile", fetchProfile, displayPage)
    a.POST("/profile", handleProfile)

    a.GET("/list", displayPage)
    a.GET("/calendar", displayPage)
  }

  r.Run(":8001")
}

func displayPage(c *gin.Context) {
  obj := gin.H{}
  obj["page"] = c.Request.URL.Path[1:]
  obj["alert"] = getSessionAlert(c)
  if user, _ := c.Get("user"); user != nil {
    obj["user"] = user
  }
  if form, _ := c.Get("form"); form != nil {
    log.Printf("=> FORM:\n   %#v\n", form) // <<< DEBUG
    obj["form"] = form
  }
  c.HTML(200, "main.tmpl", obj)
}

// --- middlewares ---

func authRequired() gin.HandlerFunc {
  return func(c *gin.Context) {
    if user := getSessionAuthInfo(c); user != nil {
      c.Set("user", user)
      return
    }
    if c.Request.URL.Path != "/" {
      setSessionAlert(c, &Alert{"warning", "You are not authorized to view this page."})
    }
    c.Redirect(302, "/login")
    c.Abort(0)
  }
}

// --- TODO list ---
// ~ display form values (password)
//   remind email - reset password
//   security: bcypt, scrypt, salt?
//   user status handling (admin/normal)
//   i18n

// --- NICE list ---
//   form validation in JS
//   pjax double render
//   proper logging
