package main

import (
  "database/sql"
  "errors"
  "log"
  "net/url"
)

type TranslationRecord struct {
  Rowid string
  Key   string
  Value string
}

type TranslationForm struct {
  Lang         string `form:"lang"`
  Key          string `form:"key"`
  Value        string `form:"value" binding:"required"`
}

func listTranslationsByQuery(q url.Values) []TranslationRecord {
  language := q.Get("language")
  if language == "" {
    language = conf.DefaultLang
  }
  return listTranslations(query["translations_lang"].Query(language))
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
    err = rows.Scan(&item.Rowid, &item.Key, &item.Value)
    if err != nil { return }
    list = append(list, item)
  }
  err = rows.Err()

  return
}

func fetchTranslation(rowid int) (TranslationForm, error) {
  var form TranslationForm
  err := query["translation_select"].QueryRow(rowid).Scan(&form.Lang, &form.Key, &form.Value)
  if err != nil {
    log.Printf("[APP] TRANSLATION-SELECT error: %s, %#v\n", err, form)
    err = errors.New("Translation was not found")
  }
  return form, err
}
