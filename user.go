package main

import (
  "database/sql"
  "errors"
  "log"
  "net/url"
  "strings"
)

type LoginForm struct {
  Email    string `form:"email"    binding:"required"`
  Password string `form:"password" binding:"required"`
}

type ForgotForm struct {
  Email    string `form:"email"    binding:"required"`
}

type UserForm struct {
  Name     string `form:"name"     binding:"required"`
  Email    string `form:"email"    binding:"required"`
  Mobile   string `form:"mobile"   binding:"required"`
  Password string `form:"password"`
  Language string `form:"language"`
  Status   int    `form:"status"`
}

type UserRecord struct {
  Id       int
  Name     string
  Email    string
  Status   int
}

func loginUserByForm(form *LoginForm) (*AuthInfo, error) {
  var auth AuthInfo
  var user_password string
  err := query["credentials_get"].QueryRow(form.Email).Scan(&user_password, &auth.Id, &auth.Name, &auth.Status, &auth.Language)
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
func sendResetLink(form *ForgotForm, lang string) bool {
  token := generateToken(16)
  res, err := query["password_forgot"].Exec(token, form.Email)
  if err != nil {
    log.Printf("[APP] RESET-FORM failure 1: %s, %s, %s\n", err, token, form.Email)
    return false
  }
  num, err := res.RowsAffected()
  if num == 0 || err != nil {
    log.Printf("[APP] RESET-FORM failure 2: %s, %s, %s\n", err, token, form.Email)
    return false
  }
  go sendResetEmail(lang, form.Email, token)
  return true
}

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

func listUsersByQuery(q url.Values) []UserRecord {
  status := q.Get("status")
  if status == "" {
    return listUsers(query["users_active"].Query())
  } else {
    return listUsers(query["users_by_status"].Query(status))
  }
}

func listUsers(rows *sql.Rows, err error) []UserRecord {
  list := []UserRecord{}
  if err != nil {
    log.Printf("[APP] USER-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item UserRecord
    err := rows.Scan(&item.Id, &item.Name, &item.Email, &item.Status)
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

func fetchUserProfile(user_id int) (UserForm, error) {
  var form UserForm
  err := query["user_select"].QueryRow(user_id).Scan(&form.Name, &form.Email, &form.Mobile, &form.Language, &form.Status)
  if err != nil {
    log.Printf("[APP] PROFILE error: %s, %#v\n", err, form)
    err = errors.New("User profile was not found")
  }
  return form, err
}

func createUser(form *UserForm) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, false, form.Language, form.Status)
  if err != nil {
    return err
  }
  form.Password = hashPassword(form.Password)
  _, err = query["user_insert"].Exec(form.Name, form.Email, form.Mobile, form.Password, form.Language, form.Status)
  if err != nil {
    log.Printf("[APP] USER-CREATE error: %s, %v\n", err, form)
    return errors.New("User could not be created. Perhaps email is already used")
  }
  return nil
}

func updateUser(user_id int, form *UserForm) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, true, form.Language, form.Status)
  if err != nil {
    return err
  }
  ok := checkFormPassword(form, user_id)
  if !ok {
    return errors.New("User could not be updated")
  }
  _, err = query["user_update"].Exec(form.Name, form.Email, form.Mobile, form.Password, form.Language, form.Status, user_id)
  if err != nil {
    log.Printf("[APP] USER-UPDATE error: %s, %d\n", err, user_id)
    return errors.New("User could not be updated. Perhaps email is already used")
  }
  return nil
}

func deleteUser(user_id int) error {
  _, err := query["user_delete"].Exec(user_id)
  if err != nil {
    log.Printf("[APP] USER-DELETE error: %s, %d\n", err, user_id)
    return errors.New("User could not be deleted")
  }
  return nil
}

func checkFormPassword(form *UserForm, user_id int) bool {
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

func validateUser(name, email, mobile, password string, allowEmpty bool, language string, status int) error {
  if !regex["name_validate"].MatchString(name) {
    return errors.New("First name and second name must be entered")
  }
  if !regex["email_validate"].MatchString(email) {
    return errors.New("Email has invalid format")
  }
  if !regex["mobile_validate"].MatchString(strings.Replace(mobile, " ", "", -1)) {
    return errors.New("Phone number has invalid format")
  }
  if len(password) < minPassLen && !(allowEmpty && password == "") {
    return errorWithA("Password must have at least %d characters", minPassLen)
  }
  if _, ok := translations[language]; !ok {
    return errors.New("Language is invalid")
  }
  if status < -2 || status > 2 {
    return errors.New("Status is invalid")
  }
  return nil
}

func userName(user_id int) (string, error) {
  var name string
  err := query["user_name"].QueryRow(user_id).Scan(&name)
  return name, err
}

func mapUserNames(user_ids []int) map[int]string {
  data := make(map[int]string, len(user_ids))
  if len(user_ids) == 0 {
    return data
  }

  rows, err := queryMultiple("users_names", user_ids)
  if err != nil {
    log.Printf("[APP] USER-NAMES error: %s\n", err)
    return data
  }
  defer rows.Close()

  var user_id int
  var name string
  for rows.Next() {
    err := rows.Scan(&user_id, &name)
    if err != nil {
      log.Printf("[APP] USER-NAMES error: %s\n", err)
    } else {
      data[user_id] = name
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] USER-NAMES error: %s\n", err)
  }
  return data
}
