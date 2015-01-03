package main

import (
  "errors"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
)

func getClassesList(c *gin.Context) {
  q := c.Request.URL.Query()
  list := listClasses(q)
  c.Set("list", list)
}

func newClassForm(c *gin.Context) {
  form := ClassForm{}
  c.Set("form", form)
}

func getClassForm(c *gin.Context) {
  var form ClassForm
  class_id, err := classId(c)
  if err == nil {
    form, err = fetchClass(class_id)
  }
  if err != nil {
    forwardWarning(c, "/classes", err.Error())
    c.Abort(0)
  } else {
    c.Set("id", class_id)
    c.Set("form", form)
  }
}

func handleClassCreate(c *gin.Context) {
  var form ClassForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  err := createClass(&form)
  if err != nil {
    showError(c, err, &form)
  } else {
    forwardTo(c, "/classes", "Class has been created")
  }
}

func handleClassUpdate(c *gin.Context) {
  var form ClassForm
  if ok := c.BindWith(&form, binding.Form); !ok {
    showError(c, errors.New("Please provide all details"), &form)
    return
  }
  class_id, err := classId(c)
  if err == nil {
    err = updateClass(class_id, &form)
  }
  if err != nil {
    showError(c, err, &form)
  } else {
    forwardTo(c, "/classes", "Class has been updated")
  }
}

func handleClassDelete(c *gin.Context) {
  class_id, err := classId(c)
  if err == nil {
    err = deleteClass(class_id)
  }
  if err != nil {
    showError(c, err)
  } else {
    forwardTo(c, "/classes", "Class has been deleted")
  }
}

// --- local helpers ---

func classId(c *gin.Context) (int, error) {
  id := c.Params.ByName("id")
  class_id, err := strconv.Atoi(id)

  if err != nil {
    return 0, errors.New("Critical error happened, please contact website admin")
  }
  return class_id, nil
}
