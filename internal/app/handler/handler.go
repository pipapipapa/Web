package handler

import (
	"orbit-calc/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", h.ShowOrbitsPage)
	router.GET("/orbit/:id", h.ShowOrbitDetailPage)
	router.GET("/mission/:id", h.ShowMissionPage) 

	router.POST("/mission/add", h.AddOrbitToMission) 
	router.POST("/mission/delete", h.DeleteMission)
}

func (h *Handler) RegisterTemplates(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}