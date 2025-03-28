package main

import (
	"github.com/labstack/echo/v4"

	"r2_example/config"
	"r2_example/controller"
	"r2_example/services/r2"
)

func main() {
	v := config.NewViper("local")

	dyLog := config.NewDyLog(v)

	r2 := r2.NewR2Services(v, dyLog)

	r2Controller := controller.NewR2Controller(dyLog, r2)

	c := echo.New()

	c.POST("/upload", r2Controller.UploadFile)
	c.GET("/lists", r2Controller.GetListsFile)
	c.POST("/public", r2Controller.MakeFilePublic)
	c.GET("/file/:id", r2Controller.GetFileByID)

	c.Start(":8080")
}
