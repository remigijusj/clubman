package main

import (
  "bytes"
  "crypto/rand"
  "errors"
  "fmt"
  "log"
  "strconv"
  "strings"
  "time"

  "code.google.com/p/go.crypto/bcrypt"
  "github.com/gin-gonic/gin"
)

type Alert struct {
  Kind, Message string
}

type AuthInfo struct {
  Id       int
  Name     string
  Status   int
  Language string
}

type ErrorWithArgs struct {
  Message  string
  Args     []interface{}
}

type SimpleRecord struct {
  Id       int
  Text     string
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

// NOTE: extra is used for optional form reference
func showError(c *gin.Context, err error, extra ...interface{}) {
  message := err.Error()
  if len(message) > 0 {
    var args []interface{}
    if erra, ok := err.(*ErrorWithArgs); ok {
      args = erra.Args
    }
    setSessionAlert(c, &Alert{"warning", TC(c, message, args...)})
  }
  if len(extra) > 0 && extra[0] != nil {
    setFlashedForm(c, extra[0])
  }
  var route string
  if len(extra) > 1 {
    route = extra[1].(string)
  } else {
    route = c.Request.URL.Path
  }
  c.Redirect(302, route)
}

func gotoSuccess(c *gin.Context, route string, message string, args ...interface{}) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"success", TC(c, message, args...)})
  }
  c.Redirect(302, route)
}

func gotoWarning(c *gin.Context, route string, message string, args ...interface{}) {
  if len(message) > 0 {
    setSessionAlert(c, &Alert{"warning", TC(c, message, args...)})
  }
  c.Redirect(302, route)
}

// --- params helpers ---

func getIntParam(c *gin.Context, name string) (int, error) {
  s := c.Params.ByName(name)
  v, err := strconv.Atoi(s)
  if err != nil {
    return 0, errors.New(panicError)
  }
  return v, nil
}

// NOTE: optional, therefore not error
func getIntQuery(c *gin.Context, name string) (int, bool) {
  s := c.Request.FormValue(name)
  v, err := strconv.Atoi(s)
  if err != nil {
    return 0, false
  }
  return v, true
}

func getDateQuery(c *gin.Context, name string) (time.Time, bool) {
  date, err := time.Parse("2006-01-02", c.Request.FormValue(name))
  if err != nil {
    return today(), false
  }
  return date, true
}

// --- session helpers ---

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

func setSavedPath(c *gin.Context, path string) {
  session, _ := cookie.Get(c.Request, "saved-path")
  defer session.Save(c.Request, c.Writer)
  session.Values["path"] = path
}

func getSavedPath(c *gin.Context) interface{} {
  session, _ := cookie.Get(c.Request, "saved-path")
  defer session.Save(c.Request, c.Writer)
  session.Options.MaxAge = -1
  return session.Values["path"]
}

func deleteSession(c *gin.Context) {
  session, _ := cookie.Get(c.Request, sessionKey)
  defer session.Save(c.Request, c.Writer)
  session.Options.MaxAge = -1
}

// --- authorization helpers ---

func currentUser(c *gin.Context) *AuthInfo {
  self, _ := c.Get("self")
  auth, ok := self.(AuthInfo)
  if ok {
    return &auth
  } else {
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
    return auth.Status == userStatusAdmin
  } else {
    return false
  }
}

func (self AuthInfo) IsAdmin() bool {
  return self.Status == userStatusAdmin
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

// --- error helpers ---

// NOTE: fmt.Sprintf no fit because we use reuslt for i18n
func (err *ErrorWithArgs) Error() string {
  return err.Message
}

func errorWithA(message string, args ...interface{}) *ErrorWithArgs {
  return &ErrorWithArgs{message, args}
}

// --- miscelaneous ---

func listRecords(name string, args ...interface{}) []SimpleRecord {
  list := []SimpleRecord{}
  rows, err := query[name].Query(args...)
  if err != nil {
    log.Printf("[APP] USER-LIST-STATUS error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item SimpleRecord
    err := rows.Scan(&item.Id, &item.Text)
    if err != nil {
      log.Printf("[APP] USER-LIST-STATUS error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-LIST error: %s\n", err)
  }
  return list
}

func containsInt(list []int, given int) bool {
  for _, val := range list {
    if val == given {
      return true
    }
  }
  return false
}

func dict(values ...interface{}) (map[string]interface{}, error) {
  if len(values)%2 != 0 {
    return nil, errors.New("invalid dict call")
  }
  hash := make(map[string]interface{}, len(values)/2)
  for i := 0; i < len(values); i += 2 {
    key, ok := values[i].(string)
    if !ok {
      return nil, errors.New("dict keys must be strings")
    }
    hash[key] = values[i+1]
  }
  return hash, nil
}

func sqlOrderById(list []int) string {
  buf := bytes.NewBufferString(" ORDER BY case id")
  buf.Grow(len(list) * 15 + 4)
  for i, id := range list {
    buf.WriteString(fmt.Sprintf(" when %d then %d", id, i))
  }
  buf.WriteString(" end")
  return buf.String()
}
