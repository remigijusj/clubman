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
    forwardTo(c, defaultPage, "")
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
  ok := sendResetLink(&form, getLang(c))
  if !ok {
    showError(c, errors.New("Reminder email could not be sent"))
  } else {
    forwardTo(c, "/login", "Email with instructions was sent to %s", form.Email)
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
  var form UserForm
  var err error
  self := currentUser(c)
  if self != nil {
    form, err = fetchUserProfile(self.Id)
  } else {
    err = errors.New("missing self")
  }
  if err != nil {
    forwardTo(c, defaultPage, "Critical error happened, please contact website admin")
    c.Abort(0)
  } else {
    c.Set("form", form)
  }
}

// NOTE: similar to handleUserUpdate
func handleProfile(c *gin.Context) {
  var form UserForm
  if ok := c.BindWith(&form, binding.Form); !ok {
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
    forwardTo(c, defaultPage, "User profile has been updated")
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
  list := listUsers(q)
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
    forwardWarning(c, "/users", err.Error())
    c.Abort(0)
  } else {
    c.Set("id", user_id)
    c.Set("form", form)
  }
}

// NOTE: serves both signup and create by admin
func handleUserCreate(c *gin.Context) {
  var form UserForm
  if ok := c.BindWith(&form, binding.Form); !ok || form.Password == "" {
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
      forwardTo(c, "/users", "User profile has been created")
    } else {
      forwardTo(c, "/login", "User profile has been created. Please wait for administrator confirmation")
    }
  }
}

// NOTE: similar to handleProfile
func handleUserUpdate(c *gin.Context) {
  var form UserForm
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
    forwardTo(c, "/users", "User profile has been updated")
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
    forwardTo(c, "/users", "User profile has been deleted")
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
