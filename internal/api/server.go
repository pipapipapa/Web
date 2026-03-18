package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {

	router := gin.Default()

	router.Static("/static", "./resources")

	router.LoadHTMLGlob("templates/*")


	if err := router.Run(":8080"); err != nil {
		logrus.Error(err)
	}
}