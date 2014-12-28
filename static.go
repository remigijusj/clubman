package main

const (
  serverHost = "nk-fitness.dk"
  serverRoot = "http://nk-fitness.dk"
  serverPort = ":8001"

  cookieHost = ""
  cookieAuth = "nk-fitness#nk-fitness#nk-fitness" // 32 bytes
  cookieEncr = "nk-fitness$nk-fitness$nk-fitness" // 32 bytes
  cookieAge  = 3600 * 1 // 1 hours
)

var queries = map[string]string{
  "credentials_get":  "SELECT password, id, name, status FROM users WHERE email=? AND status>=0",
  "password_select":  "SELECT password FROM users WHERE id=?",
  "password_forgot":  "UPDATE users SET reset_token=? WHERE email=? AND status>=0",
  "password_resets":  "SELECT id, name, status FROM users WHERE reset_token=? AND email=? AND status>=0",
  "users_active":     "SELECT id, name, email, status FROM users WHERE status>=0 ORDER BY name",
  "users_by_status":  "SELECT id, name, email, status FROM users WHERE status=? ORDER BY name",
  "user_select":      "SELECT name, email, mobile, language, status FROM users WHERE id=?",
  "user_insert":      "INSERT INTO users(name, email, mobile, password, language, status) values (?, ?, ?, ?, ?, ?)",
  "user_update":      "UPDATE users SET name=?, email=?, mobile=?, password=?, language=?, status=? WHERE id=?",
  "user_delete":      "DELETE FROM users WHERE id=?",
}
