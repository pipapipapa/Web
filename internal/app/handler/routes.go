package handler

import (
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	
	_ "orbit-calc/docs" 
)

func (h *Handler) RegisterAPI(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")

	api.GET("/orbits", h.API_GetOrbits)
	api.GET("/orbits/:id", h.API_GetOrbitDetail)
	
	api.POST("/register", h.API_Register) // Настоящий метод регистрации
	api.POST("/login", h.API_Login)       // Настоящий метод входа (возвращает JWT)


	protected := api.Group("/")
	protected.Use(h.AuthMiddleware())
	{
		protected.POST("/logout", h.API_Logout) // Добавляет токен в Redis Blacklist

		protected.POST("/mission-orbit-items", h.API_AddMissionOrbitItem)
		protected.PUT("/mission-orbit-items", h.API_UpdateMissionOrbitItem)
		protected.DELETE("/mission-orbit-items", h.API_DeleteMissionOrbitItem)

		protected.GET("/mission", h.API_GetMissionDraft)
		protected.GET("/missions", h.API_GetMissionsList)
		protected.GET("/missions/:id", h.API_GetMissionDetail) 
		protected.PUT("/missions/:id", h.API_UpdateMission)
		protected.PUT("/missions/:id/form", h.API_FormMission)
		protected.DELETE("/missions/:id", h.API_DeleteMission)

		moderatorOnly := protected.Group("/")
		moderatorOnly.Use(RoleMiddleware("MODERATOR"))
		{
			moderatorOnly.POST("/orbits", h.API_AddOrbit) 
			
			moderatorOnly.PUT("/missions/:id/complete", h.API_CompleteMission)
		}
	}
}