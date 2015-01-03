package main

import (
  "github.com/gin-gonic/gin"
)

func defineRoutes(r *gin.Engine) {
  // TODO: combine to 1 public dir
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
  r.GET("/signup", displayPage)
  r.GET("/resets",  handleReset)
  r.POST("/login",  handleLogin)
  r.POST("/forgot", handleForgot)
  r.POST("/signup", handleUserCreate)
}

func defineRoutesInternal(r *gin.Engine) {
  a := r.Group("/", authRequired())
  {
    a.GET("/", redirectToDefault)
    a.GET("/logout", handleLogout)
    a.GET("/profile", getProfile, displayPage)
    a.POST("/profile", handleProfile)

    a.GET("/classes", getClassesList, displayPage)

    a.GET("/list", displayPage)
    a.GET("/calendar", displayPage)

    defineRoutesAdmin(a)
  }
}

func defineRoutesAdmin(a *gin.RouterGroup) {
  ad := a.Group("/", adminRequired())
  {
    ad.GET("/users", getUsersList, displayPage)
    ad.GET("/users/create", newUserForm, displayPage)
    ad.POST("/users/create", handleUserCreate)
    ad.GET("/users/update/:id", getUserForm, displayPage)
    ad.POST("/users/update/:id", handleUserUpdate)
    ad.POST("/users/delete/:id", handleUserDelete)
  }
}
