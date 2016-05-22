package main

import (
  "database/sql"
  "net/url"
  "time"
)

type LogRecord struct {
  Id         int
  CreatedAt  time.Time
}

func listLogsByQuery(q url.Values, date_from time.Time) []LogRecord {
  return listLogs(query["logs_all"].Query(date_from.Format(dateFormat)))
}

func listLogs(rows *sql.Rows, err error) (list []LogRecord) {
  list = []LogRecord{}

  defer func() {
    if err != nil {
      logPrintf("LOG-LIST error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  for rows.Next() {
    var item LogRecord
    err = rows.Scan(&item.Id, &item.CreatedAt)
    if err != nil { return }
    // WARNING: see the comment in listEvents  
    item.CreatedAt = item.CreatedAt.UTC()
    list = append(list, item)
  }
  err = rows.Err()

  return
}
