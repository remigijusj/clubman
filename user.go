package main

import (
  "errors"
  "log"
  "strconv"

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

type ResetsForm struct {
  Token    string `form:"token"    binding:"required"`
  Email    string `form:"email"    binding:"required"`
  Password string `form:"password" binding:"required"`
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
  Password string `form:"password"`
  Language string `form:"language"`
}

type UserRecord struct {
  Id       int
  Name     string
  Email    string
}

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

func handleSignup(c *gin.Context) {
  var form SignupForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    displayError(c, "Please provide all details")
    return
  }
  err := createUser(&form)
  if err != nil {
    displayError(c, err.Error())
  } else {
    forwardTo(c, "/login", "Please login with the entered credentials")
  }
}

func fetchProfile(c *gin.Context) {
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

func fetchUsersList(c *gin.Context) {
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

func fetchUserProfile(c *gin.Context) {
  var form *ProfileForm
  user := c.Params.ByName("user_id")
  if user_id, err := strconv.Atoi(user); err == nil {
    form = getUser(user_id)
  }
  if form != nil {
    c.Set("page", "users_profile") // override
    c.Set("user", user)
    c.Set("form", *form)
  } else {
    forwardTo(c, "/users", "ERROR: user profile not found.") // <<< warning
    c.Abort(0)
  }
}

// --- user actions ---

func loginUser(form *LoginForm) (int, error) {
  var user_id int
  var user_password string
  err := query["credentials_get"].QueryRow(form.Email).Scan(&user_id, &user_password)
  if err != nil {
    log.Printf("[APP] LOGIN failure: %#v, %#v\n", err, form.Email)
    return 0, errors.New("Invalid password or email")
  }
  ok := comparePassword(user_password, form.Password)
  if !ok {
    return 0, errors.New("Invalid password or email")
  } else {
    return user_id, nil
  }
}

// NOTE: we don't reveal if email is missing or another problem occured
// TODO: add extensive logging
func sendResetLink(form *ForgotForm) bool {
  token := generateToken(16)
  res, err := query["password_forgot"].Exec(token, form.Email)
  if err != nil {
    return false
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    return false
  }
  go sendResetEmail(form.Email, token)
  return true
}

// TODO: maybe reset old password?
func resetLinkLogin(token, email string) (int, error) {
  var user_id int
  err := query["password_resets"].QueryRow(token, email).Scan(&user_id)
  if err == nil {
    _, err = query["password_forgot"].Exec("", email)
  }
  if err == nil {
    // TODO: internal logging
  } else {
    log.Printf("[APP] RESETS error: %s, token=%s, email=%s\n", err, token, email)
  }
  return user_id, err
}

func createUser(form *SignupForm) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, false)
  if err != nil {
    return err
  }
  form.Password = hashPassword(form.Password)
  _, err = query["user_insert"].Exec(form.Name, form.Email, form.Mobile, form.Password)
  if err != nil {
    return errors.New("User could not be created. Perhaps email is already used.")
  }
  return nil
}

func updateUser(form *ProfileForm, user_id int) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, true)
  if err != nil {
    return err
  }
  ok := checkFormPassword(form, user_id)
  if !ok {
    return errors.New("User could not be updated. Password is invalid.")
  }
  _, err = query["user_update"].Exec(form.Name, form.Email, form.Mobile, form.Password, form.Language, user_id)
  if err != nil {
    return errors.New("User could not be updated. Perhaps email is already used.")
  }
  return nil
}

func getUser(user_id int) *ProfileForm {
  var form ProfileForm
  err := query["user_select"].QueryRow(user_id).Scan(&form.Name, &form.Email, &form.Mobile, &form.Language)
  if err != nil {
    log.Printf("[APP] PROFILE error: %s, %#v\n", err, form)
    return nil
  }
  return &form
}

func validateUser(name, email, mobile, password string, allowEmpty bool) error {
  log.Printf("=> VALIDATE\n   %#v, %#v, %#v, %#v\n", name, email, mobile, password) // <<< DEBUG
  return nil
}

func checkFormPassword(form *ProfileForm, user_id int) bool {
  if form.Password != "" {
    form.Password = hashPassword(form.Password)
    if form.Password == "" {
      return false
    }
  } else {
    var currentPassword string
    err := query["password_select"].QueryRow(user_id).Scan(&currentPassword)
    if err != nil {
      return false
    }
    form.Password = currentPassword
  }
  return true
}
