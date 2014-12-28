package main

const (
  serverHost = "nk-fitness.dk"
  serverRoot = "http://nk-fitness.dk"
  serverPort = ":8001"

  cookieHost = ""
  cookieAuth = "nk-fitness#nk-fitness#nk-fitness" // 32 bytes
  cookieEncr = "nk-fitness$nk-fitness$nk-fitness" // 32 bytes
  cookieAge  = 3600 * 8 // 8 hours
)

var queries = map[string]string{
  "credentials_get": "SELECT password, id, name, status FROM users WHERE email=?",
  "password_select": "SELECT password FROM users WHERE id=?",
  "password_forgot": "UPDATE users SET reset_token=? WHERE email=?",
  "password_resets": "SELECT id, name, status FROM users WHERE reset_token=? AND email=?",
  "user_select":     "SELECT name, email, mobile, language FROM users WHERE id=?",
  "user_insert":     "INSERT INTO users(name, email, mobile, password, language) values (?, ?, ?, ?, ?)",
  "user_update":     "UPDATE users SET name=?, email=?, mobile=?, password=?, language=? WHERE id=?",
  "user_delete":     "DELETE FROM users WHERE id=?",
  "user_list":       "SELECT id, name, email FROM users",
}
