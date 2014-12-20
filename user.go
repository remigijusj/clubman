package main

import (
  "errors"
  "fmt"

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
    forwardTo(c, "/login", "User profile has been updated.")
  }
}

// --- user actions ---

func loginUser(form LoginForm) (int, error) {
  var user_id int
  var user_password string
  err := query["user_password"].QueryRow(form.Email).Scan(&user_id, &user_password)
  if err != nil {
    fmt.Printf("LOGIN FAIL: %#v, %#v\n", err, form.Email)
    return 0, errors.New("Invalid password or email")
  } else if user_password != form.Password { // <<< security!
    return 0, errors.New("Invalid password or email")
  } else {
    return user_id, nil
  }
}

func sendResetLink(form ForgotForm) error {
  fmt.Printf("RESET: %#v\n", form.Email)
  // <<< send email
  return nil
}

/*
func resetPassword() error {
  fmt.Printf("%#v\n", form.Email)
  // <<< query["user_password_reset"].Exec(form.Email)
  return nil
}
*/

func createUser(form SignupForm) error {
  fmt.Printf("CREATE: %#v, %#v, %#v, %#v\n", form.Name, form.Email, form.Mobile, form.Password)
  _, err := query["user_insert"].Exec(form.Name, form.Email, form.Mobile, form.Password)
  return err
}

func updateUser(form ProfileForm) error {
  fmt.Printf("UPDATE: %#v, %#v, %#v, %#v\n", form.Name, form.Email, form.Mobile, form.Password)
  _, err := query["user_update"].Exec(form.Name, form.Email, form.Mobile, form.Password, form.Language, form.Email)
  return err
}
