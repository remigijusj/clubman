package main

import (
  "database/sql"
  "errors"
  "log"
  "net/url"
  "strconv"
  "strings"
  "time"
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

type UserContact struct {
  Email    string
  Mobile   string
  Language string
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
func generatePasswordReset(form *ForgotForm, lang string) bool {
  password, _, err := fetchUserPassword(form.Email)
  if err != nil { return false }
  exp_unix := time.Now().Add(expireLink).Unix()
  expire := strconv.FormatInt(exp_unix, 10)
  token := computeHMAC(form.Email, expire, password)
  go sendResetLinkEmail(form.Email, lang, expire, token)
  return true
}

func verifyPasswordReset(email, expire, token string) (*AuthInfo, error) {
  password, auth, err := fetchUserPassword(email)
  if err != nil { return nil, err }
  now_unix := time.Now().Unix()
  exp_unix, err := strconv.ParseInt(expire, 10, 64)
  if err != nil { return nil, err }
  if valid := verifyHMAC(token, email, expire, password); !valid || now_unix >= exp_unix {
    err = errors.New("invalid hmac or expired")
  }
  return auth, err
}

func fetchUserPassword(email string) (string, *AuthInfo, error) {
  var auth AuthInfo
  var user_password string
  err := query["credentials_get"].QueryRow(email).Scan(&user_password, &auth.Id, &auth.Name, &auth.Status, &auth.Language)
  if err != nil {
    log.Printf("[APP] PASSWORD-FORGOT error: %s, %s, %s\n", err, email, user_password)
  }
  return user_password, &auth, err
}

func updatePassword(email, password string) error {
  if len(password) < minPassLen {
    return errorWithA("Password must have at least %d characters", minPassLen)
  }
  _, err := query["password_update"].Exec(hashPassword(password), email)
  if err != nil {
    log.Printf("[APP] PASSWORD-UPDATE error: %s, %s", err, email)
    return errors.New("User could not be updated")
  }
  return nil
}

func listUsersByQuery(q url.Values) []UserRecord {
  status := q.Get("status")
  if status == "" {
    return listUsers(query["users_active"].Query())
  } else {
    return listUsers(query["users_by_status"].Query(status))
  }
}

func listUsers(rows *sql.Rows, err error) (list []UserRecord) {
  list = []UserRecord{}

  defer func() {
    if err != nil {
      log.Printf("[APP] LIST-USERS error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  for rows.Next() {
    var item UserRecord
    err = rows.Scan(&item.Id, &item.Name, &item.Email, &item.Status)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
}

// WARNING: manual UNION-query building here, perhaps find nicer way?
func listUsersOfEventTx(tx *sql.Tx, event_id int, canceled bool) ([]UserContact, error) {
  qry, list := queries["users_of_event"], []interface{}{event_id}
  if canceled { // also notify instructor
    qry = qry + " UNION " + queries["instructor_of_event"]
    list = append(list, interface{}(event_id))
  }
  rows, err := tx.Query(qry, list...)

  return listUsersContact(rows, err), err
}

func listUsersOfEvents(event_ids []int, team_id int, canceled bool) ([]UserContact, error) {
  qry, list := multi(queries["users_of_event"], event_ids)
  if canceled { // also notify instructor
    qry = qry + " UNION " + queries["instructor_of_team"]
    list = append(list, interface{}(team_id))
  }
  rows, err := db.Query(qry, list...)

  return listUsersContact(rows, err), err
}

// NOTE: record order matches given ids order
func listUsersByIdTx(tx *sql.Tx, user_ids []int) ([]UserContact, error) {
  qry, list := multi(queries["user_contact"], user_ids)
  qry += sqlOrderById(user_ids)
  rows, err := tx.Query(qry, list...)
  return listUsersContact(rows, err), err
}

func listUsersContact(rows *sql.Rows, err error) (list []UserContact) {
  list = []UserContact{}

  defer func() {
    if err != nil {
      log.Printf("[APP] LIST-USERS-CONTACT error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  for rows.Next() {
    var item UserContact
    err = rows.Scan(&item.Email, &item.Mobile, &item.Language)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
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

// NOTE: either hash given, or take existing (if blank)
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
    return errors.New("Phone number must have format +45 12345678")
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

func fetchUserContactTx(tx *sql.Tx, user_id int) (UserContact, error) {
  var user UserContact
  err := tx.Stmt(query["user_contact"]).QueryRow(user_id).Scan(&user.Email, &user.Mobile, &user.Language)
  if err != nil {
    log.Printf("[APP] USER-CONTACT error: %s, %d, %v\n", err, user_id, user)
  }
  return user, err
}

func mapUserNames(user_ids []int) (data map[int]string) {
  data = make(map[int]string, len(user_ids))
  if len(user_ids) == 0 { return }

  var err error
  defer func() {
    if err != nil {
      log.Printf("[APP] USER-NAMES error: %s, %v\n", err, user_ids)
    }
  }()

  rows, err := multiQuery("users_names", user_ids)
  if err != nil { return }
  defer rows.Close()

  var user_id int
  var name string
  for rows.Next() {
    err = rows.Scan(&user_id, &name)
    if err != nil { return }
    data[user_id] = name
  }
  err = rows.Err()

  return
}

// --- lowercase incoming email ---

func (form *LoginForm) fixEmail() { form.Email = strings.ToLower(form.Email) }

func (form *ForgotForm) fixEmail() { form.Email = strings.ToLower(form.Email) }

func (form *UserForm) fixEmail() { form.Email = strings.ToLower(form.Email) }

type EmailFixer interface {
  fixEmail()
}
