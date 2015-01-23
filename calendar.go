package main

import (
  "time"
)

func listWeekEventsGrouped(date time.Time) [][][]EventRecord {
  from := date
  till := date.AddDate(0, 0, 7)

  data := make([][][]EventRecord, 3) // SEE: chooseWeekRow
  for w := 0; w < len(data); w++ {
    data[w] = make([][]EventRecord, 7)
    for d := 0; d < 7; d++ {
      data[w][d] = []EventRecord{}
    }
  }

  rows, err := query["events_period"].Query(from.Format(dateFormat), till.Format(dateFormat))
  list := listEvents(rows, err)

  for _, event := range list {
    d := event.StartAt.Truncate(24 * time.Hour)
    w := chooseWeekRow(event.StartAt)
    i := wdIndex(d)
    data[w][i] = append(data[w][i], event)
  }

  return data
}

func listMonthEventsGrouped(date time.Time) [][][]EventRecord {
  from := weekFirst(date)
  till := weekFirst(date.AddDate(0, 1, 0)).AddDate(0, 0, 7)
  weeks := daysDiff(from, till) / 7

  data := make([][][]EventRecord, weeks)
  for w := 0; w < len(data); w++ {
    data[w] = make([][]EventRecord, 7)
    for d := 0; d < 7; d++ {
      data[w][d] = []EventRecord{}
    }
  }

  rows, err := query["events_period"].Query(from.Format(dateFormat), till.Format(dateFormat))
  list := listEvents(rows, err)

  for _, event := range list {
    d := event.StartAt.Truncate(24 * time.Hour)
    w := daysDiff(from, d) / 7
    i := wdIndex(d)
    data[w][i] = append(data[w][i], event)
  }

  return data
}

func chooseWeekRow(t time.Time) int {
  switch h := t.Hour(); {
  case h < 12: return 0;
  case h < 16: return 1;
  default:     return 2;
  }
}

// --- datetime helpers ---

func weekFirst(d time.Time) time.Time {
  return d.Truncate(7 * 24 * time.Hour)
}

func monthFirst(d time.Time) time.Time {
  return d.AddDate(0, 0, -d.Day()+1)
}

func daysDiff(a, b time.Time) int {
  return int(b.Sub(a).Hours() / 24)
}

// with zero-sunday adjustment
func wdIndex(d time.Time) int {
  i := int(d.Weekday())
  if i > 0 {
    return i-1
  } else {
    return 6
  }
}

// used in month calendar
func calcMonthDate(date time.Time, w, d int) time.Time {
  return weekFirst(date).AddDate(0, 0, 7 * w + d)
}

func hourDuration(when time.Time) time.Duration {
  return when.Sub(when.Truncate(24 * time.Hour))
}

func today() time.Time {
  return time.Now().UTC().Truncate(24 * time.Hour)
}
