package main

import (
  "database/sql"
  "encoding/gob"
  "log"
  "regexp"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/gorilla/sessions"
  _ "github.com/mattn/go-sqlite3" // tdm-gcc
)

var db *sql.DB
var query map[string]*sql.Stmt
var regex map[string]*regexp.Regexp
var cookie *sessions.CookieStore

func init() {
  log.SetFlags(0)
  gob.Register(&Alert{})
  gob.Register(&AuthInfo{})
  gob.Register(&UserForm{})
  gob.Register(&TeamForm{})
  gob.Register(&TeamEventsForm{})
}

func main() {
  prepareQueries()
  prepareRegexes()
  prepareCookies()

  r := gin.Default()
  loadHtmlTemplates("templates/*", r)
  loadMailTemplates("mails/*")
  loadTranslations()
  makeTransHelpers()

  defineRoutes(r)
  r.Run(serverPort)
}

func prepareQueries() {
  db, _ = sql.Open("sqlite3", "./main.db")
  query = make(map[string]*sql.Stmt, len(queries))
  for name, sql := range queries {
    query[name], _ = db.Prepare(sql)
  }
}

func prepareRegexes() {
  regex = make(map[string]*regexp.Regexp, len(regexes))
  for name, rx := range regexes {
    regex[name] = regexp.MustCompile(rx)
  }
}

func prepareCookies() {
  cookie = sessions.NewCookieStore([]byte(cookieAuth), []byte(cookieEncr))
  cookie.Options = &sessions.Options{
    Domain:   cookieHost,
    Path:     "/",
    MaxAge:   cookieLife,
    HttpOnly: false,
    Secure:   false,
  }
}

func displayPage(c *gin.Context) {
  setPage(c)

  obj := gin.H(c.Keys)
  obj["alert"] = getSessionAlert(c)
  if form := getFlashedForm(c); form != nil {
    obj["form"] = form
  }
  obj["lang"] = getLang(c)
  log.Printf("=> BINDING\n   %#v\n", obj) // <<< DEBUG

  c.HTML(200, "page.tmpl", obj)
}

func authRequired() gin.HandlerFunc {
  return func(c *gin.Context) {
    if auth := getSessionAuthInfo(c); auth != nil {
      c.Set("self", *auth)
      return
    }
    if c.Request.URL.Path != "/" {
      setSessionAlert(c, &Alert{"warning", TC(c, "You are not authorized to view this page")})
    }
    c.Redirect(302, "/login")
    c.Abort(0)
  }
}

func adminRequired() gin.HandlerFunc {
  return func(c *gin.Context) {
    if isAdmin(c) {
      return
    }
    if c.Request.URL.Path != "/" {
      setSessionAlert(c, &Alert{"warning", TC(c, "You are not authorized to view this page")})
    }
    c.Redirect(302, defaultPage)
    c.Abort(0)
  }
}

func redirectToDefault(c *gin.Context) {
  c.Redirect(302, defaultPage)
}

func getLang(c *gin.Context) string {
  if lang := c.Request.URL.Query().Get("lang"); lang != "" {
    return lang
  }
  if auth := currentUser(c); auth != nil {
    return auth.Language
  }
  return defaultLang
}

func queryMultiple(name string, list []int) (*sql.Rows, error) {
  qry := strings.Replace(queries[name], "(?)", "(?"+strings.Repeat(",?", len(list)-1)+")", 1)

  args := make([]interface{}, len(list))
  for i, item := range list {
    args[i] = item
  }

  return db.Query(qry, args...)
}
