package main

import (
  "errors"
  "log"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

type LoginForm struct {
  Email    string `form:"email"    binding:"required"`
  Password string `form:"password" binding:"required"`
}

type ForgotForm struct {
  Email    string `form:"email"    binding:"required"`
}

type SignupForm struct {
  Name     string `form:"name"     binding:"required"`
  Email    string `form:"email"    binding:"required"`
  Mobile   string `form:"mobile"   binding:"required"`
  Password string `form:"password" binding:"required"`
}

type ProfileForm struct {
  Name     string `form:"name"     binding:"required"`
  Email    string `form:"email"    binding:"required"`
  Mobile   string `form:"mobile"   binding:"required"`
  Password string `form:"password" binding:"required"`
  Language string `form:"language"`
}

func handleLogin(c *gin.Context) {
  var form LoginForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please enter email and password")
    return
  }
  user_id, err := loginUser(form)
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
  err := sendResetLink(form)
  if err != nil {
    displayError(c, "Reminder email could not be sent")
  } else {
    forwardTo(c, "/login", "Please check your inbox for reminder email")
  }
}

func handleSignup(c *gin.Context) {
  var form SignupForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please provide all details")
    return
  }
  err := createUser(form)
  if err != nil {
    displayError(c, "User could not be created. Perhaps email is already used.")
  } else {
    forwardTo(c, "/login", "Please login with the entered credentials")
  }
}

func fetchProfile(c *gin.Context) {
  var form ProfileForm
  user, _ := c.Get("user");
  if user_id, ok := user.(*int); ok {
    err := query["user_select"].QueryRow(*user_id).Scan(&form.Name, &form.Email, &form.Mobile, &form.Language)
    if err != nil {
      log.Printf("[APP] PROFILE error: %s, %#v\n", err, form)
    }
  } else {
    log.Printf("[APP] PROFILE error: user=%#v\n", user)
  }
  c.Set("form", form)
}

func handleProfile(c *gin.Context) {
  var form ProfileForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please provide all details")
    return
  }
  err := updateUser(form)
  if err != nil {
    displayError(c, "User could not be updated. Perhaps email is already used.")
  } else {
    forwardTo(c, "/", "User profile has been updated.")
  }
}

// --- user actions ---

func loginUser(form LoginForm) (int, error) {
  var user_id int
  var user_password string
  err := query["password_select"].QueryRow(form.Email).Scan(&user_id, &user_password)
  if err != nil {
    log.Printf("[APP] LOGIN failure: %#v, %#v\n", err, form.Email)
    return 0, errors.New("Invalid password or email")
  } else if user_password != form.Password { // TODO: security!
    return 0, errors.New("Invalid password or email")
  } else {
    return user_id, nil
  }
}

func sendResetLink(form ForgotForm) error {
  log.Printf("=> RESET\n   %#v\n", form.Email) // <<< DEBUG
  // <<< send email
  return nil
}

/*
func resetPassword() error {
  log.Printf("%#v\n", form.Email)
  // <<< query["password_update"].Exec(form.Email)
  return nil
}
*/

func createUser(form SignupForm) error {
  log.Printf("=> CREATE\n   %#v, %#v, %#v, %#v\n", form.Name, form.Email, form.Mobile, form.Password) // <<< DEBUG
  // TODO: validate values
  _, err := query["user_insert"].Exec(form.Name, form.Email, form.Mobile, form.Password)
  return err
}

func updateUser(form ProfileForm) error {
  log.Printf("=> UPDATE\n   %#v, %#v, %#v, %#v\n", form.Name, form.Email, form.Mobile, form.Password) // <<< DEBUG
  // TODO: check password, conditionally
  _, err := query["user_update"].Exec(form.Name, form.Email, form.Mobile, form.Password, form.Language, form.Email)
  return err
}
