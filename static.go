package main

const (
  serverName = "Nykredit Fitness"

  serverHost = "demo.nk-fitness.dk"
  serverRoot = "http://demo.nk-fitness.dk"
  serverPort = ":8001"

  cookieHost = ""
  cookieAuth = "nk-fitness#nk-fitness#nk-fitness" // 32 bytes
  cookieEncr = "nk-fitness$nk-fitness$nk-fitness" // 32 bytes
  cookieLife = 3600 * 1 // 1 hours

  emailsHost = "smtp.gmail.com"
  emailsUser = ""
  emailsPass = ""
  emailsPort = 587
  emailsFrom = ""

  sessionKey = "session"
  bcryptCost = 10
  minPassLen = 6
  defaultLang = "da"
  defaultPage = "/calendar" // not "/"
)

var languages = []string{"da", "en"}

var regexes = map[string]string{
  "name_validate":   `\pL\s+\pL`,
  "email_validate":  `^\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}$`,
  "mobile_validate": `^\+?\d{8,11}$`,
}

var queries = map[string]string{
  "translations":     "SELECT locale, key, value FROM translations ORDER BY locale, key",

  "credentials_get":  "SELECT password, id, name, status, language FROM users WHERE email=? AND status>=0",
  "password_select":  "SELECT password FROM users WHERE id=?",
  "password_forgot":  "UPDATE users SET reset_token=? WHERE email=? AND status>=0",
  "password_resets":  "SELECT id, name, status FROM users WHERE reset_token=? AND email=? AND status>=0",
  "users_active":     "SELECT id, name, email, status FROM users WHERE status>=0 ORDER BY name",
  "users_by_status":  "SELECT id, name, email, status FROM users WHERE status=? ORDER BY name",
  "user_select":      "SELECT name, email, mobile, language, status FROM users WHERE id=?",
  "user_insert":      "INSERT INTO users(name, email, mobile, password, language, status) values (?, ?, ?, ?, ?, ?)",
  "user_update":      "UPDATE users SET name=?, email=?, mobile=?, password=?, language=?, status=? WHERE id=?",
  "user_delete":      "DELETE FROM users WHERE id=?",

  "classes_all":      "SELECT id, name FROM classes ORDER BY name",
  "class_select":     "SELECT name, part_min, part_max FROM classes WHERE id=?",
  "class_insert":     "INSERT INTO classes(name, part_min, part_max) VALUES (?, ?, ?)",
  "class_update":     "UPDATE classes SET name=?, part_min=?, part_max=? WHERE id=?",
  "class_delete":     "DELETE FROM classes WHERE id=?",
}
