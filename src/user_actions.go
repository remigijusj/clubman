package main

import (
  "errors"
  "strconv"

  "github.com/gin-gonic/gin"
)

func handleLogin(c *gin.Context) {
  var form LoginForm
  if ok := bindForm(c, &form); !ok {
    showError(c, errors.New("Please enter email and password"))
    return
  }
  auth, err := loginUserByForm(&form)
  if err != nil {
    showError(c, err)
  } else {
    setSessionAuthInfo(c, auth)
    gotoSuccess(c, conf.DefaultPage, "")
  }
}

func handleLogout(c *gin.Context) {
  deleteSession(c)
  gotoSuccess(c, "/login", "")
}

func handleForgot(c *gin.Context) {
  var form ForgotForm
  if ok := bindForm(c, &form); !ok {
    showError(c, errors.New("Please enter your email"))
    return
  }
  ok := generatePasswordReset(&form, getLang(c))
  if !ok {
    showError(c, errors.New("Reminder email could not be sent"))
  } else {
    gotoSuccess(c, "/login", "Email with instructions was sent to %s", form.Email)
  }
}

func getResetInfo(c *gin.Context) {
  q := c.Request.URL.Query()
  auth, err := verifyPasswordReset(q.Get("email"), q.Get("expire"), q.Get("token"))
  if err != nil {
    gotoWarning(c, "/login", "Password reset request is invalid or expired")
    c.Abort()
  } else {
    c.Set("path", c.Request.URL.String())
    c.Set("form", auth)
  }
}

func handleReset(c *gin.Context) {
  auth, err := verifyPasswordReset(c.Request.FormValue("email"), c.Request.FormValue("expire"), c.Request.FormValue("token"))
  if err != nil {
    gotoWarning(c, "/login", "Password reset request is invalid or expired")
    return
  }
  err = updatePassword(c.Request.FormValue("email"), c.Request.FormValue("password"))
  if err != nil {
    showError(c, err)
  } else {
    setSessionAuthInfo(c, auth)
    getSavedPath(c) // to avoid early redirect
    gotoSuccess(c, conf.DefaultPage, "")
  }
}

func getProfile(c *gin.Context) {
  var form UserForm
  var err error
  self := currentUser(c)
  if self != nil {
    form, err = fetchUserProfile(self.Id)
  } else {
    err = errors.New("missing self")
  }
  if err != nil {
    gotoWarning(c, conf.DefaultPage, panicError)
    c.Abort()
  } else {
    c.Set("form", form)
  }
}

// NOTE: similar to handleUserUpdate
func handleProfile(c *gin.Context) {
  var form UserForm
  if ok := bindForm(c, &form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  var err error
  self := currentUser(c)
  if self != nil {
    form.Status = self.Status // security override
    err = updateUser(self.Id, &form)
    if err == nil {
      updateSesssionNow(c, self, &form)
    }
  } else {
    err = errors.New("Critical error happened, please contact website admin")
  }
  if err != nil {
    showError(c, err, &form)
  } else {
    gotoSuccess(c, conf.DefaultPage, "User profile has been updated")
  }
}

func updateSesssionNow(c *gin.Context, self *AuthInfo, form *UserForm) {
  self.Name = form.Name
  self.Language = form.Language
  setSessionAuthInfo(c, self)
  c.Set("self", *self)
}

func getUserList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listUsersByQuery(q)
  c.Set("list", list)
}

func newUserForm(c *gin.Context) {
  form := UserForm{}
  c.Set("form", form)
}

func getUserForm(c *gin.Context) {
  var form UserForm
  user_id, err := anotherUserId(c)
  if err == nil {
    form, err = fetchUserProfile(user_id)
  }
  if err != nil {
    gotoWarning(c, "/users", err.Error())
    c.Abort()
  } else {
    c.Set("id", user_id)
    c.Set("form", form)
  }
}

// NOTE: serves both signup and create by admin
func handleUserCreate(c *gin.Context) {
  var form UserForm
  if ok := bindForm(c, &form); !ok || form.Password == "" {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  if c.Request.URL.Path == "/signup" {
    form.Language = getLang(c)
    form.Status = userStatusWaiting
  }
  err := createUser(&form)
  if err != nil {
    showError(c, err, &form)
  } else {
    if isAuthenticated(c) {
      gotoSuccess(c, "/users", "User profile has been created")
    } else {
      gotoSuccess(c, "/login", "User profile has been created. Please wait for administrator confirmation")
    }
  }
}

// NOTE: similar to handleProfile
func handleUserUpdate(c *gin.Context) {
  var form UserForm
  if ok := bindForm(c, &form); !ok {
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
    gotoSuccess(c, "/users", "User profile has been updated")
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
    gotoSuccess(c, "/users", "User profile has been deleted")
  }
}

// --- local helpers ---

func anotherUserId(c *gin.Context) (int, error) {
  id := c.Params.ByName("id")
  user_id, err := strconv.Atoi(id)
  self := currentUser(c)

  if err != nil || self == nil {
    return 0, errors.New("Critical error happened, please contact website admin")
  }
  if user_id == self.Id {
    return 0, errors.New("No access to your own profile")
  }
  return user_id, nil
}
