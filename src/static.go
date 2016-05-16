package main

var regexes = map[string]string{
  "name_validate":   `\pL\s+\pL`,
  "email_validate":  `^\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}$`,
  "mobile_validate": `^\+?\d{8,11}$`,
}

var queries = map[string]string{
  "translations":       "SELECT lang, key, value FROM translations ORDER BY lang, key",

  "credentials_get":    "SELECT password, id, name, status, language FROM users WHERE email=? AND status>=0",
  "password_select":    "SELECT password FROM users WHERE id=?",
  "password_update":    "UPDATE users SET password=? WHERE email=? AND status>=0",

  "users_active":       "SELECT id, name, email, status FROM users WHERE status>=0 ORDER BY name",
  "users_by_status":    "SELECT id, name, email, status FROM users WHERE status=? ORDER BY name",
  "users_on_status":    "SELECT id, name FROM users WHERE status=? ORDER BY name",
  "users_on_active":    "SELECT id, name FROM users WHERE status>=0 ORDER BY name",
  "users_names":        "SELECT id, name FROM users WHERE id IN (?)",
  "user_select":        "SELECT name, email, mobile, language, status FROM users WHERE id=?",
  "user_insert":        "INSERT INTO users(name, email, mobile, password, language, status) values (?, ?, ?, ?, ?, ?)",
  "user_update":        "UPDATE users SET name=?, email=?, mobile=?, password=?, language=?, status=? WHERE id=?",
  "user_delete":        "DELETE FROM users WHERE id=?",
  "user_name":          "SELECT name FROM users WHERE id=?",
  "user_contact":       "SELECT email, mobile, language FROM users WHERE id IN (?)",
  "users_of_event":     "SELECT DISTINCT email, mobile, language FROM assignments LEFT JOIN users ON user_id=users.id WHERE event_id IN (?)",
  "instructor_of_event":"SELECT DISTINCT email, mobile, language FROM events LEFT JOIN teams ON team_id=teams.id LEFT JOIN users ON instructor_id=users.id WHERE events.id IN (?)",
  "instructor_of_team": "SELECT email, mobile, language FROM teams LEFT JOIN users ON instructor_id=users.id WHERE teams.id=?",

  "teams_all":          "SELECT teams.id, teams.name, users.name, users_min, users_max FROM teams LEFT JOIN users ON teams.instructor_id=users.id ORDER BY teams.name",
  "team_names_all":     "SELECT id, name FROM teams ORDER BY name",
  "team_select":        "SELECT name, users_min, users_max, instructor_id FROM teams WHERE id=?",
  "team_users_max":     "SELECT users_max FROM events LEFT JOIN teams ON team_id=teams.id WHERE events.id=?",
  "team_insert":        "INSERT INTO teams(name, users_min, users_max, instructor_id) VALUES (?, ?, ?, ?)",
  "team_update":        "UPDATE teams SET name=?, users_min=?, users_max=?, instructor_id=? WHERE id=?",
  "team_delete":        "DELETE FROM teams WHERE id=?",

  "events_team":        "SELECT id, team_id, start_at, minutes, status FROM events WHERE team_id=? AND start_at >= ? ORDER BY start_at",
  "events_period":      "SELECT id, team_id, start_at, minutes, status FROM events WHERE start_at >= ? AND start_at < ? AND status>=0 ORDER BY start_at",
  "events_multi":       "SELECT id, start_at FROM events WHERE team_id=? AND start_at >= ? AND start_at < ?",
  "events_under":       "SELECT events.id FROM events LEFT JOIN teams ON team_id=teams.id LEFT JOIN assignments ON event_id=events.id WHERE start_at >= ? and start_at < ? GROUP BY events.id HAVING count(user_id) < users_min",
  "event_select_info":  "SELECT events.id, name, start_at, minutes, events.status FROM events LEFT JOIN teams ON team_id=teams.id WHERE events.id=?",
  "event_select":       "SELECT team_id, start_at, minutes, status FROM events WHERE id=?",
  "event_insert":       "INSERT INTO events(team_id, start_at, minutes, status) VALUES (?, ?, ?, ?)",
  "event_update":       "UPDATE events SET team_id=?, start_at=?, minutes=?, status=? WHERE id=?",
  "event_status":       "UPDATE events SET status=? WHERE id=?",
  "event_delete":       "DELETE FROM events WHERE id IN (?)",
  "events_clear":       "DELETE FROM events WHERE team_id=?",

  "assignments_event":  "SELECT user_id, users.name, assignments.status FROM assignments JOIN users ON user_id=users.id WHERE event_id=? ORDER BY assignments.status DESC, assignments.id",
  "assignments_user":   "SELECT event_id, teams.name, start_at, minutes, events.status, assignments.status FROM assignments JOIN events ON event_id=events.id JOIN teams ON team_id=teams.id WHERE user_id=? AND events.start_at >= ? ORDER BY start_at",
  "assignments_status": "SELECT event_id, status FROM assignments WHERE event_id IN (?) AND user_id=?",
  "assignments_counts": "SELECT event_id, count(id) FROM assignments WHERE event_id IN (?) GROUP BY event_id",
  "assignments_count":  "SELECT count(id) FROM assignments WHERE event_id=?",
  "assignments_queue":  "SELECT id, user_id, status FROM assignments WHERE event_id=? ORDER BY status DESC, id",
  "assignment_status":  "SELECT id, status FROM assignments WHERE event_id=? AND user_id=?",
  "assignment_insert":  "INSERT INTO assignments(event_id, user_id, status) VALUES (?, ?, ?)",
  "assignment_update":  "UPDATE assignments SET status=? WHERE event_id=? AND user_id IN (?)",
  "assignment_delete":  "DELETE FROM assignments WHERE event_id=? AND user_id=?",
  "assignments_clear":  "DELETE FROM assignments WHERE event_id IN (?)",

  "logs_all":           "SELECT id, created_at FROM logs WHERE created_at >= ?",
}