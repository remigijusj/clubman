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
  serverHost = "nk-fitness.dk"
  serverRoot = "http://nk-fitness.dk"
  cookieHost = ""
  cookieAuth = "nk-fitness#nk-fitness#nk-fitness" // 32 bytes
  cookieEncr = "nk-fitness$nk-fitness$nk-fitness" // 32 bytes
)

var queries = map[string]string{
  "credentials_get": "SELECT id, password FROM users WHERE email=?",
  "password_select": "SELECT password FROM users WHERE id=?",
  "password_forgot": "UPDATE users SET reset_token=? WHERE email=?",
  "password_resets": "SELECT id FROM users WHERE reset_token=? AND email=?",
  "user_select":     "SELECT name, email, mobile, language FROM users WHERE id=?",
  "user_insert":     "INSERT INTO users(name, email, mobile, password) values (?, ?, ?, ?)",
  "user_update":     "UPDATE users SET name=?, email=?, mobile=?, password=?, language=? WHERE id=?",
}

var query map[string]*sql.Stmt
var cookie *sessions.CookieStore

func init() {
  log.SetFlags(0)
  gob.Register(&Alert{})
}

func prepareQueries() {
  db, _ := sql.Open("sqlite3", "./main.db")
  query = make(map[string]*sql.Stmt, len(queries))
  for name, sql := range queries {
    query[name], _ = db.Prepare(sql)
  }
}

func prepareCookies() {
  cookie = sessions.NewCookieStore([]byte(cookieAuth), []byte(cookieEncr))
  cookie.Options = &sessions.Options{
    Domain:   cookieHost,
    Path:     "/",
    MaxAge:   3600 * 8, // 8 hours
    HttpOnly: false,
    Secure:   false,
  }
}

func main() {
  prepareQueries()
  prepareCookies()

  r := gin.Default()
  r.LoadHTMLFiles("main.tmpl")

  defineRoutes(r)
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

func defineRoutes(r *gin.Engine) {
  // TODO: combine to 1 public dir
  s := r.Group("/")
  {
    s.Handlers = s.Handlers[:1] // removing Logger from Default
    s.Static("/css", "./css")
    s.Static("/img", "./img")
    s.Static("/js", "./js")
  }

  r.GET("/login",  displayPage)
  r.GET("/signup", displayPage)
  r.GET("/forgot", displayPage)
  r.GET("/resets",  handleReset)
  r.POST("/login",  handleLogin)
  r.POST("/signup", handleSignup)
  r.POST("/forgot", handleForgot)

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
}
