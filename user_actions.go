package main

import (
  "errors"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func handleLogin(c *gin.Context) {
  var form LoginForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please enter email and password"))
    return
  }
  auth, err := loginUserByForm(&form)
  if err != nil {
    showError(c, err)
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
    showError(c, errors.New("Please enter your email"))
    return
  }
  ok := sendResetLink(&form)
  if !ok {
    showError(c, errors.New("Reminder email could not be sent"))
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
    forwardTo(c, "/profile", "Please enter new password and click Save")
  }
}

func getProfile(c *gin.Context) {
  var form ProfileForm
  var err error
  if self := currentUser(c); self != nil {
    form, err = fetchUserProfile(self.Id)
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
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  var err error
  if self := currentUser(c); self != nil {
    form.Status = self.Status // security override
    err = updateUser(self.Id, &form)
    // NICE: update name immediately
    if err == nil && self.Name != form.Name {
      self.Name = form.Name
      setSessionAuthInfo(c, self)
    }
  }
  if err != nil {
    showError(c, err, &form)
  } else {
    forwardTo(c, "/", "User profile has been updated.")
  }
}

func getUsersList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listUsers(q)
  c.Set("list", list)
}

func newUserForm(c *gin.Context) {
  form := ProfileForm{}
  c.Set("form", form)
}

func getUserForm(c *gin.Context) {
  var form ProfileForm
  user_id, err := anotherUserId(c)
  if err == nil {
    form, err = fetchUserProfile(user_id)
  }
  if err != nil {
    forwardWarning(c, "/users", err.Error())
    c.Abort(0)
  } else {
    c.Set("user", user_id)
    c.Set("form", form)
  }
}

// NOTE: serves both signup and create by admin
func handleUserCreate(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok || form.Password == "" {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  err := createUser(&form)
  if err != nil {
    showError(c, err, &form)
  } else {
    if isAuthenticated(c) {
      forwardTo(c, "/users", "User profile has been created.")
    } else {
      forwardTo(c, "/login", "User profile has been created. Please wait for administrator confirmation.")
    }
  }
}

// NOTE: similar to handleProfile
func handleUserUpdate(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  user_id, err := anotherUserId(c)
  if err == nil {
    err = updateUser(user_id, &form)
  }
  if err != nil {
    showError(c, err, &form)
  } else {
    forwardTo(c, "/users", "User profile has been updated.")
  }
}

func handleUserDelete(c *gin.Context) {
  user_id, err := anotherUserId(c)
  if err == nil {
    err = deleteUser(user_id)
  }
  if err != nil {
    showError(c, err)
  } else {
    forwardTo(c, "/users", "User profile has been deleted.")
  }
}

// --- local helpers ---

func anotherUserId(c *gin.Context) (int, error) {
  user := c.Params.ByName("id")
  user_id, err := strconv.Atoi(user)
  self := currentUser(c)

  if err != nil || self == nil {
    return 0, errors.New("Critical error happened. Please contact website admin.")
  }
  if user_id == self.Id {
    return 0, errors.New("No access to your own profile.")
  }
  return user_id, nil
}
