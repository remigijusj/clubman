package main

import (
  "database/sql"
  "errors"
  "net/url"
  "strings"
)

type TranslationRecord struct {
  Rowid string
  Key   string
  Value string
}

type TranslationForm struct {
  Lang         string `form:"lang"  binding:"required"`
  Key          string `form:"key"   binding:"required"`
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
      logPrintf("TRANSLATION-LIST error: %s\n", err)
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
    logPrintf("TRANSLATION-SELECT error: %s, %#v\n", err, form)
    err = errors.New("Translation was not found")
  }
  return form, err
}

func updateTranslation(rowid int, form *TranslationForm) error {
  form.Value = strings.Trim(form.Value, " ")
  err := checkTranslation(rowid, form)
  if err != nil {
    return err
  }
  _, err = query["translation_update"].Exec(form.Value, rowid)
  if err != nil {
    logPrintf("TRANSLATION-UPDATE error: %s, %d\n", err, rowid)
    return errors.New("Translation could not be updated")
  }
  replaceTranslation(form.Lang, form.Key, form.Value)
  return nil
}

func checkTranslation(rowid int, form *TranslationForm) error {
  saved, err := fetchTranslation(rowid)
  if err != nil {
    return err
  }
  if form.Value == "" {
    return errors.New("Translation cannot be empty")
  }
  if form.Lang != saved.Lang || form.Key != saved.Key {
    return errors.New("Version mismatch with the saved record")
  }
  if !equalPlaceholders(form.Key, form.Value) {
    return errors.New("Placeholders must match with the English string")
  }
  return nil
}

func equalPlaceholders(one, two string) bool {
  ones := regex["string_placeholder"].FindAllString(one, -1)
  twos := regex["string_placeholder"].FindAllString(two, -1)
  if len(ones) != len(twos) {
    return false
  }
  for i, ph := range ones {
    if ph != twos[i] {
      return false
    }
  }
  return true
}
