package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")
	
	// УСЛУГИ (3)
	api.GET("/orbits", h.API_GetOrbits)
	api.GET("/orbits/:id", h.API_GetOrbitDetail)
	api.POST("/orbits", h.API_AddOrbit)

	// М-М (3)
	api.POST("/mission-orbit-items", h.API_AddM2M)
	api.PUT("/mission-orbit-items", h.API_UpdateM2M)
	api.DELETE("/mission-orbit-items", h.API_DeleteM2M)

	// ЗАЯВКИ (7)
	api.GET("/missions/", h.API_GetDraft)
	api.GET("/missions", h.API_GetMissionsList)
	api.GET("/missions/:id", h.API_GetMissionDetail)
	api.PUT("/missions/:id", h.API_UpdateMission)
	api.PUT("/missions/:id/form", h.API_FormMission)
	api.PUT("/missions/:id/complete", h.API_CompleteMission)
	api.DELETE("/missions/:id", h.API_DeleteMission)

	// ЮЗЕРЫ (3 - Заглушки)
	api.POST("/register", func(c *gin.Context){ c.JSON(200, gin.H{"msg": "Registered"}) })
	api.POST("/login", func(c *gin.Context){ c.JSON(200, gin.H{"msg": "Logged in"}) })
	api.POST("/logout", func(c *gin.Context){ c.JSON(200, gin.H{"msg": "Logged out"}) })
}