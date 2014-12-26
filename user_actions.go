package main

import (
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
  auth, err := loginUserByForm(&form)
  if err != nil {
    displayError(c, err.Error())
  } else {
    setSessionAuthInfo(c, auth)
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
  auth, err := loginUserByToken(q.Get("token"), q.Get("email"))
  if err != nil {
    forwardWarning(c, "/login", "Password reset request is invalid or expired")
  } else {
    setSessionAuthInfo(c, auth)
    forwardTo(c, "/profile", "Please enter new password and click save")
  }
}

func getProfile(c *gin.Context) {
  var form ProfileForm
  var err error
  if user_id, ok := currentUserId(c); ok {
    form, err = fetchUserProfile(user_id)
  }
  if err != nil {
    forwardTo(c, "/", "Critical error happened. Please contact website admin.")
    c.Abort(0)
  } else {
    c.Set("form", form)
  }
}

// NOTE: similar to handleUserUpdate
func handleProfile(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please provide all details")
    return
  }
  var err error
  if user_id, ok := currentUserId(c); ok {
    err = updateUser(&form, user_id)
  }
  if err != nil {
    displayError(c, err.Error())
  } else {
    forwardTo(c, "/", "User profile has been updated.")
  }
}

// TODO: filter, search
func getUsersList(c *gin.Context) {
  list := listUsers()
  c.Set("list", list)
}

func newUserForm(c *gin.Context) {
  form := ProfileForm{}
  c.Set("form", form)
}

func getUserForm(c *gin.Context) {
  var form ProfileForm
  user := c.Params.ByName("id")
  user_id, err := strconv.Atoi(user)
  if err == nil {
    form, err = fetchUserProfile(user_id)
  }
  if err != nil {
    forwardWarning(c, "/users", "ERROR: user profile not found.")
    c.Abort(0)
  } else {
    c.Set("user", user_id)
    c.Set("form", form)
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

// NOTE: similar to handleProfile
func handleUserUpdate(c *gin.Context) {
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
