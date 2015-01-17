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
  rows, err := query["teams_all"].Query()
  return listTeams(rows, err)
}

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
  _, err = query["events_delete_team"].Exec(team_id)
  if err != nil {
    log.Printf("[APP] EVENTS-DELETE error: %s, %d\n", err, team_id)
    return nil
  }
  return nil
}

func validateTeam(name string, users_min, users_max, instructor_id int) error {
  if users_min < 0 || users_max < 0 || (users_max > 0 && users_max < users_min) {
    return errors.New("Participant numbers are invalid")
  }
  return nil
}

func teamName(team_id int) (string, error) {
  var name string
  err := query["team_name"].QueryRow(team_id).Scan(&name)
  return name, err
}
