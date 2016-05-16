package main

import (
  "database/sql"
  "encoding/gob"
  "log"
  "os"
  "regexp"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/gorilla/sessions"
  _ "github.com/mattn/go-sqlite3" // tdm-gcc
  "github.com/robfig/cron"
)

var conf *Conf
var db *sql.DB
var query map[string]*sql.Stmt
var regex map[string]*regexp.Regexp
var clock *cron.Cron
var cookie *sessions.CookieStore

func init() {
  log.SetFlags(0)
  gob.Register(&Alert{})
  gob.Register(&AuthInfo{})
  gob.Register(&UserForm{})
  gob.Register(&TeamForm{})
  gob.Register(&TeamEventsForm{})
  gob.Register(&EventForm{})
}

func main() {
  setGinMode()
  prepareConfig()

  prepareQueries()
  prepareRegexes()
  prepareCookies()

  startCronService()

  r := gin.Default()
  loadHtmlTemplates("templates/*", r)
  loadMailTemplates("mails/*")
  loadTranslations()
  makeTransHelpers()

  defineRoutes(r)
  r.Run(conf.ServerPort)
}

func setGinMode() {
  if mode := os.Getenv(gin.ENV_GIN_MODE); mode == "" {
    gin.SetMode(gin.ReleaseMode)
  }
}

func debugMode() bool {
  return gin.Mode() == gin.DebugMode
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
  cookie = sessions.NewCookieStore([]byte(conf.CookieAuth), []byte(conf.CookieEncr))
  cookie.Options = &sessions.Options{
    Domain:   conf.CookieHost,
    Path:     "/",
    MaxAge:   int(conf.CookieLife.Seconds()),
    HttpOnly: false,
    Secure:   false,
  }
}

func startCronService() {
  clock = cron.New()
  clock.AddFunc(conf.CancelCheck, autoCancelEvents)
  clock.Start()
}

func displayPage(c *gin.Context) {
  setPage(c)

  obj := gin.H(c.Keys)
  obj["alert"] = getSessionAlert(c)
  if form := getFlashedForm(c); form != nil {
    obj["form"] = form
  }
  obj["lang"] = getLang(c)
  if debugMode() {
    log.Printf("=> BINDING\n   %#v\n", obj)
  }

  c.HTML(200, "page.tmpl", obj)
}

func authRequired() gin.HandlerFunc {
  return func(c *gin.Context) {
    if auth := getSessionAuthInfo(c); auth != nil {
      c.Set("self", *auth)
      if path, ok := getSavedPath(c).(string); ok {
        c.Redirect(302, path)
        c.Abort()
      }
    } else {
      if c.Request.URL.Path != "/" {
        setSessionAlert(c, &Alert{"warning", TC(c, permitError)})
      }
      setSavedPath(c, c.Request.URL.Path)
      c.Redirect(302, "/login")
      c.Abort()
    }
  }
}

func adminRequired() gin.HandlerFunc {
  return func(c *gin.Context) {
    if isAdmin(c) {
      return
    }
    if c.Request.URL.Path != "/" {
      setSessionAlert(c, &Alert{"warning", TC(c, permitError)})
    }
    c.Redirect(302, conf.DefaultPage)
    c.Abort()
  }
}

func redirectToDefault(c *gin.Context) {
  c.Redirect(302, conf.DefaultPage)
}

func getLang(c *gin.Context) string {
  if lang := c.Request.URL.Query().Get("lang"); lang != "" {
    return lang
  }
  if auth := currentUser(c); auth != nil {
    return auth.Language
  }
  return conf.DefaultLang
}

func multiQuery(name string, args ...interface{}) (*sql.Rows, error) {
  qry, list := multi(queries[name], args...)
  return db.Query(qry, list...)
}

func multiExec(name string, args ...interface{}) (sql.Result, error) {
  qry, list := multi(queries[name], args...)
  return db.Exec(qry, list...)
}

// expands int slice arguments for IN-queries
// WARNING: caller is reponsible for ensuring non-empty slices
func multi(qry string, args ...interface{}) (string, []interface{}) {
  list := []interface{}{}
  for _, item := range args {
    if ints, ok := item.([]int); ok {
      for _, it := range ints {
        list = append(list, it)
      }
      // WARNING: a hack, only works once!
      qry = strings.Replace(qry, "(?)", "(?"+strings.Repeat(",?", len(ints)-1)+")", 1)
    } else {
      list = append(list, item)
    }
  }
  return qry, list
}
