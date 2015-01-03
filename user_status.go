package main

type UserStatus struct {
  Status   int
  Title    string
}

const (
  userStatusWaiting = -1
  userStatusAdmin = 2
)

var statuses = []UserStatus{
  UserStatus{-2, "Inactive"     },
  UserStatus{-1, "Waiting"      },
  UserStatus{ 0, "User"         },
  UserStatus{ 1, "Instructor"   },
  UserStatus{ 2, "Administrator"},
}

func statusTitle(status int) string {
  for _, us := range statuses {
    if us.Status == status {
      return us.Title
    }
  }
  return ""
}

func statusList() []UserStatus {
  return statuses
}
