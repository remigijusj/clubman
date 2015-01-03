package main

import (
  "database/sql"
  "log"
  "net/url"
)

type ClassRecord struct {
  Id       int
  Name     string
  PartMin  int
  PartMax  int
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
