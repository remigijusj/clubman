package main

const (
  configFile  = "config.toml"
  configVar   = "CONFIG_TOML"

  panicError  = "Critical error happened, please contact website admin"
  permitError = "You are not authorized to view this page"

  timeFormat  = "15:04"
  dateFormat  = "2006-01-02" // db
  fullFormat  = "2006-01-02 15:04:05" // db
  logsPrefix  = "[APP] 2006/01/02 - 15:04:05 | "
)

var regexes = map[string]string{
  "name_validate":      `\pL\s+\pL`,
  "email_validate":     `^\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}$`,
  "mobile_validate":    `^\+?\d{8,11}$`,
  "string_placeholder": `%[sd]`,
  "query_placeholder": `\(\$\d+\)|\$\d+`,
}

// NOTE: ($\d) placeholder must be last in the query 
var queries = map[string]string{
  "translations":       "SELECT lang, key, value FROM translations ORDER BY lang, key",
  "translations_lang":  "SELECT id, key, value FROM translations WHERE lang=$1 ORDER BY value",
  "translation_select": "SELECT lang, key, value FROM translations WHERE id=$1",
  "translation_update": "UPDATE translations SET value=$1 WHERE id=$2",

  "credentials_get":    "SELECT password, id, name, status, language FROM users WHERE email=$1 AND status>=0",
  "password_select":    "SELECT password FROM users WHERE id=$1",
  "password_update":    "UPDATE users SET password=$1 WHERE email=$2 AND status>=0",

  "users_active":       "SELECT id, name, email, status FROM users WHERE status>=0 ORDER BY name",
  "users_by_status":    "SELECT id, name, email, status FROM users WHERE status=$1 ORDER BY name",
  "users_on_status":    "SELECT id, name FROM users WHERE status=$1 ORDER BY name",
  "users_on_active":    "SELECT id, name FROM users WHERE status>=0 ORDER BY name",
  "users_names":        "SELECT id, name FROM users WHERE id IN ($1)",
  "user_select":        "SELECT name, email, mobile, language, status FROM users WHERE id=$1",
  "user_insert":        "INSERT INTO users(name, email, mobile, password, language, status) values ($1, $2, $3, $4, $5, $6)",
  "user_update":        "UPDATE users SET name=$1, email=$2, mobile=$3, password=$4, language=$5, status=$6 WHERE id=$7",
  "user_delete":        "DELETE FROM users WHERE id=$1",
  "user_name":          "SELECT name FROM users WHERE id=$1",
  "user_contact":       "SELECT email, mobile, language FROM users WHERE id IN ($1)",
  "users_of_event":     "SELECT DISTINCT email, mobile, language FROM assignments LEFT JOIN users ON user_id=users.id WHERE event_id IN ($1)",
  "instructor_of_event":"SELECT DISTINCT email, mobile, language FROM events LEFT JOIN teams ON team_id=teams.id LEFT JOIN users ON instructor_id=users.id WHERE events.id=$1",
  "instructor_of_team": "SELECT email, mobile, language FROM teams LEFT JOIN users ON instructor_id=users.id WHERE teams.id=$1",
  "user_teams_count":   "SELECT count(id) FROM teams WHERE instructor_id=$1",

  "teams_all":          "SELECT teams.id, teams.name, users.name, users_min, users_max FROM teams LEFT JOIN users ON teams.instructor_id=users.id ORDER BY teams.name",
  "team_names_all":     "SELECT id, name FROM teams ORDER BY name",
  "team_select":        "SELECT name, users_min, users_max, instructor_id FROM teams WHERE id=$1",
  "team_users_max":     "SELECT users_max FROM events LEFT JOIN teams ON team_id=teams.id WHERE events.id=$1",
  "team_insert":        "INSERT INTO teams(name, users_min, users_max, instructor_id) VALUES ($1, $2, $3, $4)",
  "team_update":        "UPDATE teams SET name=$1, users_min=$2, users_max=$3, instructor_id=$4 WHERE id=$5",
  "team_delete":        "DELETE FROM teams WHERE id=$1",

  "events_team":        "SELECT id, team_id, start_at, minutes, status FROM events WHERE team_id=$1 AND start_at >= $2 ORDER BY start_at",
  "events_period":      "SELECT id, team_id, start_at, minutes, status FROM events WHERE start_at >= $1 AND start_at < $2 ORDER BY start_at",
  "events_multi":       "SELECT id, start_at FROM events WHERE team_id=$1 AND start_at >= $2 AND start_at < $3",
  "events_under":       "SELECT events.id FROM events LEFT JOIN teams ON team_id=teams.id LEFT JOIN assignments ON event_id=events.id WHERE start_at >= $1 and start_at < $2 AND events.status=0 GROUP BY events.id, users_min HAVING count(user_id) < users_min",
  "event_select_info":  "SELECT events.id, name, start_at, minutes, events.status FROM events LEFT JOIN teams ON team_id=teams.id WHERE events.id=$1",
  "event_select":       "SELECT team_id, start_at, minutes, status FROM events WHERE id=$1",
  "event_insert":       "INSERT INTO events(team_id, start_at, minutes, status) VALUES ($1, $2, $3, $4)",
  "event_update":       "UPDATE events SET team_id=$1, start_at=$2, minutes=$3, status=$4 WHERE id=$5",
  "event_status":       "UPDATE events SET status=$1 WHERE id=$2",
  "event_delete":       "DELETE FROM events WHERE id IN ($1)",
  "events_clear":       "DELETE FROM events WHERE team_id=$1",

  "assignments_event":  "SELECT user_id, users.name, assignments.status FROM assignments JOIN users ON user_id=users.id WHERE event_id=$1 ORDER BY assignments.status DESC, assignments.id",
  "assignments_user":   "SELECT event_id, teams.name, start_at, minutes, events.status, assignments.status FROM assignments JOIN events ON event_id=events.id JOIN teams ON team_id=teams.id WHERE user_id=$1 AND events.start_at >= $2 ORDER BY start_at",
  "assignments_status": "SELECT event_id, status FROM assignments WHERE user_id=$1 AND event_id IN ($2)",
  "assignments_counts": "SELECT event_id, count(id) FROM assignments WHERE event_id IN ($1) GROUP BY event_id",
  "assignments_count":  "SELECT count(id) FROM assignments WHERE event_id=$1",
  "assignments_queue":  "SELECT id, user_id, status FROM assignments WHERE event_id=$1 ORDER BY status DESC, id",
  "assignment_status":  "SELECT id, status FROM assignments WHERE event_id=$1 AND user_id=$2",
  "assignment_insert":  "INSERT INTO assignments(event_id, user_id, status) VALUES ($1, $2, $3)",
  "assignment_update":  "UPDATE assignments SET status=$1 WHERE event_id=$2 AND user_id IN ($3)",
  "assignment_delete":  "DELETE FROM assignments WHERE event_id=$1 AND user_id=$2",
  "assignments_clear":  "DELETE FROM assignments WHERE event_id IN ($1)",
  "assignments_prune":  "DELETE FROM assignments WHERE user_id=$1",

  "logs_all":           "SELECT id, created_at FROM logs WHERE created_at >= $1",
}
