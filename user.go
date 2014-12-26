package main

import (
  "errors"
  "log"
)

type LoginForm struct {
  Email    string `form:"email"    binding:"required"`
  Password string `form:"password" binding:"required"`
}

type ForgotForm struct {
  Email    string `form:"email"    binding:"required"`
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

func loginUserByForm(form *LoginForm) (*AuthInfo, error) {
  var auth AuthInfo
  var user_password string
  err := query["credentials_get"].QueryRow(form.Email).Scan(&user_password, &auth.Id, &auth.Name, &auth.Status)
  if err != nil {
    log.Printf("[APP] LOGIN-FORM failure: %s, %s\n", err, form.Email)
    return nil, errors.New("Invalid password or email")
  }
  ok := comparePassword(user_password, form.Password)
  if !ok {
    return nil, errors.New("Invalid password or email")
  } else {
    return &auth, nil
  }
}

// NOTE: we don't reveal if email is missing or another problem occured
// TODO: add extensive logging
func sendResetLink(form *ForgotForm) bool {
  token := generateToken(16)
  res, err := query["password_forgot"].Exec(token, form.Email)
  if err != nil {
    log.Printf("[APP] RESET-FORM failure: %s, %s, %s\n", err, token, form.Email)
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
func loginUserByToken(token, email string) (*AuthInfo, error) {
  var auth AuthInfo
  err := query["password_resets"].QueryRow(token, email).Scan(&auth.Id, &auth.Name, &auth.Status)
  if err == nil {
    _, err = query["password_forgot"].Exec("", email)
  }
  if err != nil {
    log.Printf("[APP] LOGIN-TOKEN error: %s, token=%s, email=%s\n", err, token, email)
  }
  return &auth, err
}

func listUsers() []UserRecord {
  list := []UserRecord{}
  rows, err := query["user_list"].Query()
  if err != nil {
    log.Printf("[APP] USER-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item UserRecord
    err := rows.Scan(&item.Id, &item.Name, &item.Email)
    if err != nil {
      log.Printf("[APP] USER-LIST error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-LIST error: %s\n", err)
  }
  return list
}

func createUser(form *ProfileForm) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, false)
  if err != nil {
    return err
  }
  form.Password = hashPassword(form.Password)
  _, err = query["user_insert"].Exec(form.Name, form.Email, form.Mobile, form.Password, form.Language)
  if err != nil {
    log.Printf("[APP] USER-CREATE error: %s, %v\n", err, form)
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
    log.Printf("[APP] USER-UPDATE error: %s, %d\n", err, user_id)
    return errors.New("User could not be updated. Perhaps email is already used.")
  }
  return nil
}

func deleteUser(user_id int) error {
  _, err := query["user_delete"].Exec(user_id)
  if err != nil {
    log.Printf("[APP] USER-DELETE error: %s, %d\n", err, user_id)
    return errors.New("User could not be deleted.")
  }
  return nil
}

func fetchUserProfile(user_id int) (ProfileForm, error) {
  var form ProfileForm
  err := query["user_select"].QueryRow(user_id).Scan(&form.Name, &form.Email, &form.Mobile, &form.Language)
  if err != nil {
    log.Printf("[APP] PROFILE error: %s, %#v\n", err, form)
  }
  return form, err
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
      log.Printf("[APP] PASSWORD-SELECT error: %s, %d\n", err, user_id)
      return false
    }
    form.Password = currentPassword
  }
  return true
}

func validateUser(name, email, mobile, password string, allowEmpty bool) error {
  log.Printf("=> VALIDATE\n   %#v, %#v, %#v, %#v\n", name, email, mobile, password) // <<< DEBUG
  // TODO: implement
  return nil
}
