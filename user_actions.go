package main

import (
  "errors"
  "log"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func handleLogin(c *gin.Context) {
  var form LoginForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please enter email and password")
    return
  }
  user_id, err := loginUser(&form)
  if err != nil {
    displayError(c, err.Error())
  } else {
    setSessionAuthInfo(c, user_id)
    forwardTo(c, "/", "")
  }
}

func handleLogout(c *gin.Context) {
  deleteSession(c)
  forwardTo(c, "/login", "")
}

func handleForgot(c *gin.Context) {
  var form ForgotForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please enter your email")
    return
  }
  ok := sendResetLink(&form)
  if !ok {
    displayError(c, "Reminder email could not be sent")
  } else {
    forwardTo(c, "/login", "Email with instructions was sent to "+form.Email)
  }
}

func handleReset(c *gin.Context) {
  q := c.Request.URL.Query()
  user_id, err := resetLinkLogin(q.Get("token"), q.Get("email"))
  if err == nil {
    setSessionAuthInfo(c, user_id)
    forwardTo(c, "/profile", "Please enter new password and click save")
  } else {
    forwardTo(c, "/login", "Password reset request is invalid or expired") // TODO: warning
  }
}

// TODO: eliminate *, error from getUser
func getProfile(c *gin.Context) {
  var form *ProfileForm
  if user_id, ok := currentUserId(c); ok {
    form = getUser(user_id)
  }
  if form != nil {
    c.Set("form", *form)
  } else {
    forwardTo(c, "/", "Critical error happened. Please contact website admin.")
    c.Abort(0)
  }
}

func handleProfile(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please provide all details")
    return
  }
  err := errors.New("")
  if user_id, ok := currentUserId(c); ok {
    err = updateUser(&form, user_id)
  }
  if err != nil {
    displayError(c, err.Error())
  } else {
    forwardTo(c, "/", "User profile has been updated.")
  }
}

// TODO: refactor to user.go (query)
func getUsersList(c *gin.Context) {
  rows, err := query["user_list"].Query()
  if err != nil {
    log.Printf("[APP] USER_LIST error: %s\n", err)
    return
  }
  defer rows.Close()
  list := []UserRecord{}
  for rows.Next() {
    var item UserRecord
    err := rows.Scan(&item.Id, &item.Name, &item.Email)
    if err != nil {
      log.Printf("[APP] USER_LIST error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER_LIST error: %s\n", err)
  }
  c.Set("list", list)
}

func newUserForm(c *gin.Context) {
  c.Set("form", ProfileForm{})
}

// TODO: eliminate *, error from getUser
func getUserForm(c *gin.Context) {
  var form *ProfileForm
  user := c.Params.ByName("id")
  user_id, err := strconv.Atoi(user)
  if err == nil {
    form = getUser(user_id)
  }
  if form != nil {
    c.Set("user", user_id)
    c.Set("form", *form)
  } else {
    forwardWarning(c, "/users", "ERROR: user profile not found.")
    c.Abort(0)
  }
}

func handleUserCreate(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok || form.Password == "" {
    displayError(c, "Please provide all details")
    return
  }
  err := createUser(&form)
  if err != nil {
    displayError(c, err.Error())
  } else {
    if isAuthenticated(c) {
      forwardTo(c, "/users", "User has been created.")
    } else {
      forwardTo(c, "/login", "Please login with the entered credentials")
    }
  }
}

// TODO: de-duplicate with handleProfile
func handleUserProfile(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please provide all details")
    return
  }
  var err error
  user := c.Params.ByName("id")
  if user_id, err := strconv.Atoi(user); err == nil {
    err = updateUser(&form, user_id)
  }
  if err != nil {
    displayError(c, err.Error())
  } else {
    forwardTo(c, "/users", "User profile has been updated.")
  }
}

func handleUserDelete(c *gin.Context) {
  user := c.Params.ByName("id")
  user_id, err := strconv.Atoi(user)
  self_id, ok := currentUserId(c)
  if err != nil || !ok {
    displayError(c, "Critical error happened. Please contact website admin.")
    return
  }
  if user_id == self_id {
    displayError(c, "You can't delete own profile.")
    return
  }
  err = deleteUser(user_id)
  if err != nil {
    displayError(c, err.Error())
  } else {
    forwardTo(c, "/users", "User has been deleted.")
  }
}
