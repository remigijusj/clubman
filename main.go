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

var query map[string]*sql.Stmt
var cookie *sessions.CookieStore

func init() {
  log.SetFlags(0)
  gob.Register(&Alert{})
  gob.Register(&AuthInfo{})
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
  setPage(c)

  obj := gin.H(c.Keys)
  obj["alert"] = getSessionAlert(c)
  log.Printf("=> BINDING\n   %#v\n", obj) // <<< DEBUG

  c.HTML(200, "main.tmpl", obj)
}

func authRequired() gin.HandlerFunc {
  return func(c *gin.Context) {
    if auth := getSessionAuthInfo(c); auth != nil {
      c.Set("self", *auth)
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
  r.GET("/forgot", displayPage)
  r.GET("/signup", displayPage)
  r.GET("/resets",  handleReset)
  r.POST("/login",  handleLogin)
  r.POST("/forgot", handleForgot)
  r.POST("/signup", handleUserCreate)

  a := r.Group("/", authRequired())
  {
    // TODO: redirect to welcome?
    a.GET("/", displayPage)
    a.GET("/logout", handleLogout)
    a.GET("/profile", getProfile, displayPage)
    a.POST("/profile", handleProfile)

    // TODO: admin group
    a.GET("/users", getUsersList, displayPage)
    a.GET("/users/create", newUserForm, displayPage)
    a.POST("/users/create", handleUserCreate)
    a.GET("/users/update/:id", getUserForm, displayPage)
    a.POST("/users/update/:id", handleUserUpdate)
    a.POST("/users/delete/:id", handleUserDelete)

    // TODO: implement
    a.GET("/list", displayPage)
    a.GET("/calendar", displayPage)
  }
}
