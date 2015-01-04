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
  "translations":     "SELECT lang, key, value FROM translations ORDER BY lang, key",

  "credentials_get":  "SELECT password, id, name, status, language FROM users WHERE email=? AND status>=0",
  "password_select":  "SELECT password FROM users WHERE id=?",
  "password_forgot":  "UPDATE users SET reset_token=? WHERE email=? AND status>=0",
  "password_resets":  "SELECT id, name, status FROM users WHERE reset_token=? AND email=? AND status>=0",

  "users_active":     "SELECT id, name, email, status FROM users WHERE status>=0 ORDER BY name",
  "users_by_status":  "SELECT id, name, email, status FROM users WHERE status=? ORDER BY name",
  "users_on_status":  "SELECT id, name FROM users WHERE status=? ORDER BY name",
  "user_select":      "SELECT name, email, mobile, language, status FROM users WHERE id=?",
  "user_insert":      "INSERT INTO users(name, email, mobile, password, language, status) values (?, ?, ?, ?, ?, ?)",
  "user_update":      "UPDATE users SET name=?, email=?, mobile=?, password=?, language=?, status=? WHERE id=?",
  "user_delete":      "DELETE FROM users WHERE id=?",
  "user_name":        "SELECT name FROM users WHERE id=?",

  "teams_all":        "SELECT id, name FROM teams ORDER BY name",
  "team_select":      "SELECT name, part_min, part_max, instructor_id FROM teams WHERE id=?",
  "team_insert":      "INSERT INTO teams(name, part_min, part_max, instructor_id) VALUES (?, ?, ?, ?)",
  "team_update":      "UPDATE teams SET name=?, part_min=?, part_max=?, instructor_id=? WHERE id=?",
  "team_delete":      "DELETE FROM teams WHERE id=?",
  "team_events":      "SELECT id, start_at, minutes, status FROM events WHERE team_id=? ORDER BY start_at",
}
