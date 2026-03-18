package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")

	api.GET("/orbits", h.API_GetOrbits)
	api.GET("/orbits/:id", h.API_GetOrbitDetail)
	api.POST("/orbits", h.API_AddOrbit)

	api.POST("/mission-orbit-items", h.API_AddMissionOrbitItem)
	api.PUT("/mission-orbit-items", h.API_UpdateMissionOrbitItem)
	api.DELETE("/mission-orbit-items", h.API_DeleteMissionOrbitItem)

	api.GET("/missions/", h.API_GetMissionDraft)
	api.GET("/mission", h.API_GetMissionsList)
	api.GET("/missions/:id", h.API_GetMissionDetail)
	api.PUT("/missions/:id", h.API_UpdateMission)
	api.PUT("/missions/:id/form", h.API_FormMission)
	api.PUT("/missions/:id/complete", h.API_CompleteMission)
	api.DELETE("/missions/:id", h.API_DeleteMission)

	api.POST("/register", func(c *gin.Context) { c.JSON(200, gin.H{"msg": "Registered"}) })
	api.POST("/login", func(c *gin.Context) { c.JSON(200, gin.H{"msg": "Logged in"}) })
	api.POST("/logout", func(c *gin.Context) { c.JSON(200, gin.H{"msg": "Logged out"}) })
}
