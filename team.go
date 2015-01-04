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
}

type TeamForm struct {
  Name         string `form:"name"          binding:"required"`
  PartMin      int    `form:"part_min"`
  PartMax      int    `form:"part_max"`
  InstructorId int    `form:"instructor_id" binding:"required"`
}

func listTeams(q url.Values) []TeamRecord {
  list := []TeamRecord{}
  rows, err := listTeamsQuery(q)
  if err != nil {
    log.Printf("[APP] TEAM-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item TeamRecord
    err := rows.Scan(&item.Id, &item.Name)
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

func listTeamsQuery(q url.Values) (*sql.Rows, error) {
  return query["teams_all"].Query()
}

func fetchTeam(team_id int) (TeamForm, error) {
  var form TeamForm
  err := query["team_select"].QueryRow(team_id).Scan(&form.Name, &form.PartMin, &form.PartMax, &form.InstructorId)
  if err != nil {
    log.Printf("[APP] TEAM-SELECT error: %s, %#v\n", err, form)
    err = errors.New("Team was not found")
  }
  return form, err
}

func createTeam(form *TeamForm) error {
  err := validateTeam(form.Name, form.PartMin, form.PartMax, form.InstructorId)
  if err != nil {
    return err
  }
  _, err = query["team_insert"].Exec(form.Name, form.PartMin, form.PartMax, form.InstructorId)
  if err != nil {
    log.Printf("[APP] TEAM-CREATE error: %s, %v\n", err, form)
    return errors.New("Team could not be created")
  }
  return nil
}

func updateTeam(team_id int, form *TeamForm) error {
  err := validateTeam(form.Name, form.PartMin, form.PartMax, form.InstructorId)
  if err != nil {
    return err
  }
  _, err = query["team_update"].Exec(form.Name, form.PartMin, form.PartMax, form.InstructorId, team_id)
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
  return nil
}

func validateTeam(name string, part_min, part_max, instructor_id int) error {
  if part_min < 0 || part_max < 0 || (part_max > 0 && part_max < part_min) {
    return errors.New("Participant numbers are invalid")
  }
  return nil
}
