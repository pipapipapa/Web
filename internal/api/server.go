package api

import (
	"orbit-calc/internal/app/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	h := &handler.Handler{}

	router := gin.Default()

	router.Static("/static", "./resources")

	router.LoadHTMLGlob("templates/*")

	router.GET("/", h.GetOrbits)
	router.GET("/orbit/:id", h.GetOrbitDetail)
	router.GET("/mission/:id", h.GetOrbitMission)

	if err := router.Run(":8080"); err != nil {
		logrus.Error(err)
	}
}