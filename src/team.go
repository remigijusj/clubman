package main

import (
  "database/sql"
  "errors"
  "net/url"
)

type TeamRecord struct {
  Id       int
  Name     string
  UserName string
  UsersMin int
  UsersMax int
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

func listTeams(rows *sql.Rows, err error) (list []TeamRecord) {
  list = []TeamRecord{}

  defer func() {
    if err != nil {
      logPrintf("TEAM-LIST error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  for rows.Next() {
    var item TeamRecord
    err = rows.Scan(&item.Id, &item.Name, &item.UserName, &item.UsersMin, &item.UsersMax)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
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

func fetchTeam(team_id int) (TeamForm, error) {
  var form TeamForm
  err := query["team_select"].QueryRow(team_id).Scan(&form.Name, &form.UsersMin, &form.UsersMax, &form.InstructorId)
  if err != nil {
    logPrintf("TEAM-SELECT error: %s, %#v\n", err, form)
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
    logPrintf("TEAM-INSERT error: %s, %v\n", err, form)
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
    logPrintf("TEAM-UPDATE error: %s, %d\n", err, team_id)
    return errors.New("Team could not be updated")
  }
  return nil
}

func deleteTeam(team_id int) error {
  _, err := query["team_delete"].Exec(team_id)
  if err != nil {
    logPrintf("TEAM-DELETE error: %s, %d\n", err, team_id)
    return errors.New("Team could not be deleted")
  }
  err = clearEvents(team_id)
  if err != nil {
    logPrintf("TEAM-DELETE-ASSIGNMENTS error: %s, %d\n", err, team_id)
    return errors.New("Team assignments could not be deleted")
  }
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
    logPrintf("TEAM-USERS-MAX error: %s, %d\n", err, event_id)
  }
  return max, err
}
