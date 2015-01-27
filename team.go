package main

import (
  "database/sql"
  "errors"
  "log"
  "net/url"
)

type TeamRecord struct {
  Id       int
  Name     string
  UserName string
}

type TeamForm struct {
  Name         string `form:"name"          binding:"required"`
  UsersMin     int    `form:"users_min"`
  UsersMax     int    `form:"users_max"`
  InstructorId int    `form:"instructor_id" binding:"required"`
}

func listTeamsByQuery(q url.Values) []TeamRecord {
  return listTeams(query["teams_all"].Query())
}

// NOTE: with more teams, we could pass event_ids to select
func indexTeams() map[int]TeamRecord {
  rows, err := query["teams_all"].Query()
  list := listTeams(rows, err)
  data := make(map[int]TeamRecord, len(list))
  for _, team := range list {
    data[team.Id] = team
  }
  return data
}

func listTeams(rows *sql.Rows, err error) []TeamRecord {
  list := []TeamRecord{}
  if err != nil {
    log.Printf("[APP] TEAM-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item TeamRecord
    err := rows.Scan(&item.Id, &item.Name, &item.UserName)
    if err != nil {
      log.Printf("[APP] TEAM-LIST error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] TEAM-LIST error: %s\n", err)
  }
  return list
}

func fetchTeam(team_id int) (TeamForm, error) {
  var form TeamForm
  err := query["team_select"].QueryRow(team_id).Scan(&form.Name, &form.UsersMin, &form.UsersMax, &form.InstructorId)
  if err != nil {
    log.Printf("[APP] TEAM-SELECT error: %s, %#v\n", err, form)
    err = errors.New("Team was not found")
  }
  return form, err
}

func createTeam(form *TeamForm) error {
  err := validateTeam(form.Name, form.UsersMin, form.UsersMax, form.InstructorId)
  if err != nil {
    return err
  }
  _, err = query["team_insert"].Exec(form.Name, form.UsersMin, form.UsersMax, form.InstructorId)
  if err != nil {
    log.Printf("[APP] TEAM-CREATE error: %s, %v\n", err, form)
    return errors.New("Team could not be created")
  }
  return nil
}

func updateTeam(team_id int, form *TeamForm) error {
  err := validateTeam(form.Name, form.UsersMin, form.UsersMax, form.InstructorId)
  if err != nil {
    return err
  }
  _, err = query["team_update"].Exec(form.Name, form.UsersMin, form.UsersMax, form.InstructorId, team_id)
  if err != nil {
    log.Printf("[APP] TEAM-UPDATE error: %s, %d\n", err, team_id)
    return errors.New("Team could not be updated")
  }
  return nil
}

func deleteTeam(team_id int) error {
  _, err := query["team_delete"].Exec(team_id)
  if err != nil {
    log.Printf("[APP] TEAM-DELETE error: %s, %d\n", err, team_id)
    return errors.New("Team could not be deleted")
  }
  clearEvents(team_id)
  return nil
}

func validateTeam(name string, users_min, users_max, instructor_id int) error {
  if users_min < 0 || users_max < 0 || (users_max > 0 && users_max < users_min) {
    return errors.New("Participant numbers are invalid")
  }
  return nil
}

func maxTeamUsersTx(tx *sql.Tx, event_id int) (int, error) {
  var max int
  err := tx.Stmt(query["team_users_max"]).QueryRow(event_id).Scan(&max)
  if err != nil {
    log.Printf("[APP] TEAM-USERS-MAX error: %s, %d\n", err, event_id)
  }
  return max, err
}
