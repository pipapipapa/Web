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

// RegisterRoutes - регистрирует все маршруты приложения
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", h.ShowOrbitsPage)
	router.GET("/orbit/:id", h.ShowOrbitDetailPage)
	router.GET("/mission/:id", h.ShowMissionPage) 

	router.POST("/mission/add", h.AddOrbitToMission) 
	router.POST("/mission/delete", h.DeleteMission)
	router.POST("/mission/remove-item", h.RemoveItem)
	router.POST("/mission/form", h.FormMission)
	router.POST("/mission/complete", h.CompleteMission)
}

// RegisterTemplates - регистрирует статику и шаблоны
func (h *Handler) RegisterTemplates(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}