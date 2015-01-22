package main

import (
  "github.com/gin-gonic/gin"
)

func defineRoutes(r *gin.Engine) {
  s := r.Group("/")
  {
    s.Handlers = s.Handlers[:1] // removing Logger from Default
    s.Static("/css", "./css")
    s.Static("/img", "./img")
    s.Static("/js", "./js")
  }
  defineRoutesExternal(r)
  defineRoutesInternal(r)
}

func defineRoutesExternal(r *gin.Engine) {
  r.GET("/login",  displayPage)
  r.GET("/forgot", displayPage)
  //r.GET("/signup", displayPage)
  r.GET("/resets",  handleReset)
  r.POST("/login",  handleLogin)
  r.POST("/forgot", handleForgot)
  //r.POST("/signup", handleUserCreate)
}

func defineRoutesInternal(r *gin.Engine) {
  a := r.Group("/", authRequired())
  {
    a.GET("/", redirectToDefault)
    a.GET("/logout", handleLogout)
    a.GET("/profile", getProfile, displayPage)
    a.POST("/profile", handleProfile)

    a.GET("/assignments",     getSelfAssignmentsList,          displayPage)
    a.GET("/users/view/:id",  getUserAssignmentsList,          displayPage)

    a.GET("/teams",           getTeamList,                     displayPage)
    a.GET("/teams/view/:id",  getTeamForm,  getTeamEventsData, displayPage)

    a.GET("/events/view/:id", getEventForm, getEventTeam, getEventAssignments, displayPage)

    a.GET("/calendar",        redirectCalendar)
    a.GET("/calendar/week",   getWeekData,                     displayPage)
    a.GET("/calendar/month",  getMonthData,                    displayPage)

    a.POST("/assignments/create/:event_id", handleAssignmentCreate)
    a.POST("/assignments/delete/:event_id", handleAssignmentDelete)

    defineRoutesAdmin(a)
  }
}

func defineRoutesAdmin(a *gin.RouterGroup) {
  ad := a.Group("/", adminRequired())
  {
    ad.GET("/users",              getUserList, displayPage)
    ad.GET("/users/create",       newUserForm, displayPage)
    ad.GET("/users/update/:id",   getUserForm, displayPage)
    ad.POST("/users/create",      handleUserCreate)
    ad.POST("/users/update/:id",  handleUserUpdate)
    ad.POST("/users/delete/:id",  handleUserDelete)

    ad.GET("/teams/create",       newTeamForm, displayPage)
    ad.GET("/teams/update/:id",   getTeamForm, displayPage)
    ad.GET("/teams/events/:id",   getTeamForm, getTeamEventsData, newTeamEventsForm, displayPage)
    ad.POST("/teams/create",      handleTeamCreate)
    ad.POST("/teams/update/:id",  handleTeamUpdate)
    ad.POST("/teams/delete/:id",  handleTeamDelete)

    ad.GET("/events/update/:id",  getEventForm, displayPage)
    ad.POST("/events/update/:id", handleEventUpdate)
    ad.POST("/events/delete/:id", handleEventDelete)
    ad.POST("/events/create/:team_id", handleEventsCreate) // NOTE: ?-> add
    ad.POST("/events/cancel/:team_id", handleEventsCancel)
    ad.POST("/events/remove/:team_id", handleEventsRemove)
  }
}
