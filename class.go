package main

import (
  "database/sql"
  "errors"
  "log"
  "net/url"
)

type ClassRecord struct {
  Id       int
  Name     string
}

type ClassForm struct {
  Name     string `form:"name"     binding:"required"`
  PartMin  int    `form:"part_min"`
  PartMax  int    `form:"part_max"`
}

func listClasses(q url.Values) []ClassRecord {
  list := []ClassRecord{}
  rows, err := listClassesQuery(q)
  if err != nil {
    log.Printf("[APP] CLASS-LIST error: %s\n", err)
    return list
  }
  defer rows.Close()
  for rows.Next() {
    var item ClassRecord
    err := rows.Scan(&item.Id, &item.Name)
    if err != nil {
      log.Printf("[APP] CLASS-LIST error: %s\n", err)
    } else {
      list = append(list, item)
    }
  }
  if err := rows.Err(); err != nil {
    log.Printf("[APP] CLASS-LIST error: %s\n", err)
  }
  return list
}

func listClassesQuery(q url.Values) (*sql.Rows, error) {
  return query["classes_all"].Query()
}

func fetchClass(class_id int) (ClassForm, error) {
  var form ClassForm
  err := query["class_select"].QueryRow(class_id).Scan(&form.Name, &form.PartMin, &form.PartMax)
  if err != nil {
    log.Printf("[APP] CLASS-SELECT error: %s, %#v\n", err, form)
    err = errors.New("Class was not found")
  }
  return form, err
}

func createClass(form *ClassForm) error {
  err := validateClass(form.Name, form.PartMin, form.PartMax)
  if err != nil {
    return err
  }
  _, err = query["class_insert"].Exec(form.Name, form.PartMin, form.PartMax)
  if err != nil {
    log.Printf("[APP] CLASS-CREATE error: %s, %v\n", err, form)
    return errors.New("Class could not be created")
  }
  return nil
}

func updateClass(class_id int, form *ClassForm) error {
  err := validateClass(form.Name, form.PartMin, form.PartMax)
  if err != nil {
    return err
  }
  _, err = query["class_update"].Exec(form.Name, form.PartMin, form.PartMax, class_id)
  if err != nil {
    log.Printf("[APP] CLASS-UPDATE error: %s, %d\n", err, class_id)
    return errors.New("Class could not be updated")
  }
  return nil
}

func deleteClass(class_id int) error {
  _, err := query["class_delete"].Exec(class_id)
  if err != nil {
    log.Printf("[APP] CLASS-DELETE error: %s, %d\n", err, class_id)
    return errors.New("Class could not be deleted")
  }
  return nil
}

func validateClass(name string, part_min, part_max int) error {
  if part_min < 0 || part_max < 0 || (part_max > 0 && part_max < part_min) {
    return errors.New("Participant numbers are invalid")
  }
  return nil
}
