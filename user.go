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

type ProfileForm struct {
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

type UserStatus struct {
  Status   int
  Title    string
}

const userStatusAdmin = 2

var statuses = []UserStatus{
  UserStatus{-2, "Deactivated"  },
  UserStatus{-1, "Waiting"      },
  UserStatus{ 0, "User"         },
  UserStatus{ 1, "Instructor"   },
  UserStatus{ 2, "Administrator"},
}

func statusTitle(status int) string {
  for _, us := range statuses {
    if us.Status == status {
      return us.Title
    }
  }
  return ""
}

func statusList() []UserStatus {
  return statuses
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

func listUsers(q url.Values) []UserRecord {
  list := []UserRecord{}
  rows, err := listUsersQuery(q)
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

func listUsersQuery(q url.Values) (*sql.Rows, error) {
  status := q.Get("status")
  if status == "" {
    return query["users_active"].Query()
  } else {
    return query["users_by_status"].Query(status)
  }
}

func fetchUserProfile(user_id int) (ProfileForm, error) {
  var form ProfileForm
  err := query["user_select"].QueryRow(user_id).Scan(&form.Name, &form.Email, &form.Mobile, &form.Language, &form.Status)
  if err != nil {
    log.Printf("[APP] PROFILE error: %s, %#v\n", err, form)
    err = errors.New("User profile was not found")
  }
  return form, err
}

func createUser(form *ProfileForm) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, false, form.Status)
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

func updateUser(user_id int, form *ProfileForm) error {
  err := validateUser(form.Name, form.Email, form.Mobile, form.Password, true, form.Status)
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

func validateUser(name, email, mobile, password string, allowEmpty bool, status int) error {
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
    return errorWithA("Password must have at least %{cnt} characters", minPassLen)
  }
  if status < -2 || status > 2 {
    return errors.New("Status is invalid")
  }
  return nil
}
