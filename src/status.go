package main

type Status struct {
  Status   int
  Title    string
}

const (
  userStatusWaiting    = -1
  userStatusInstructor = 1
  userStatusAdmin      = 2
  eventStatusDeleted     = -3 // just for internal signaling
  eventStatusCanceled    = -2
  eventStatusActive      = 0
  assignmentStatusWaiting   = -2
  assignmentStatusNotified  = -1
  assignmentStatusConfirmed = 1
)

var statuses = map[string][]Status {
  "user": []Status{
    Status{-2, "Inactive"     },
    Status{-1, "Waiting"      },
    Status{ 0, "User"         },
    Status{ 1, "Instructor"   },
    Status{ 2, "Administrator"},
  },
  "event": []Status{
    Status{-2, "Canceled" },
    Status{ 0, "Active"   },
  },
  "assignment": []Status{
    Status{-2, "Waiting"              },
    Status{-1, "Waiting confirmation" },
    Status{ 1, "Confirmed"            },
  },
}

func statusTitle(kind string, status int) string {
  for _, us := range statuses[kind] {
    if us.Status == status {
      return us.Title
    }
  }
  return ""
}

func statusList(kind string) []Status {
  return statuses[kind]
}
