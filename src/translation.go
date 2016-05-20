package main

import (
  "database/sql"
  "log"
  "net/url"
)

type TranslationRecord struct {
  Lang  string
  Key   string
  Value string
}

func listTranslationsByQuery(q url.Values) []TranslationRecord {
  return listTranslations(query["translations"].Query())
}

func listTranslations(rows *sql.Rows, err error) (list []TranslationRecord) {
  list = []TranslationRecord{}

  defer func() {
    if err != nil {
      log.Printf("[APP] TRANSLATION-LIST error: %s\n", err)
    }
  }()
  if err != nil { return }

  defer rows.Close()

  for rows.Next() {
    var item TranslationRecord
    err = rows.Scan(&item.Lang, &item.Key, &item.Value)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
}
