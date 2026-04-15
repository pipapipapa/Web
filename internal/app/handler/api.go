package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"orbit-calc/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// Услуги

// API_GetOrbits godoc
// @Summary Получить список услуг (орбит)
// @Description Доступно всем (Гость). Возвращает список активных орбит.
// @Tags Orbits
// @Produce json
// @Param search query string false "Фильтр по названию"
// @Success 200 {object} map[string]interface{}
// @Router /api/orbits [get]
func (h *Handler) API_GetOrbits(c *gin.Context) {
	q := c.Query("search")
	orbits, _ := h.Repository.GetOrbits(q)
	c.JSON(200, gin.H{"data": orbits})
}

// API_GetOrbitDetail godoc
// @Summary Детальная информация об орбите
// @Description Доступно всем (Гость).
// @Tags Orbits
// @Produce json
// @Param id path int true "ID орбиты"
// @Success 200 {object} map[string]interface{}
// @Router /api/orbits/{id} [get]
func (h *Handler) API_GetOrbitDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var orbit ds.OrbitType
	err := h.Repository.DB().First(&orbit, id).Error
	if err != nil {
		c.JSON(404, gin.H{"error": "Услуга не найдена"})
		return
	}

	c.JSON(200, gin.H{"data": orbit})
}

// API_AddOrbit godoc
// @Summary Добавить новую орбиту (ТОЛЬКО МОДЕРАТОР)
// @Security ApiKeyAuth
// @Tags Orbits
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Название"
// @Param description formData string true "Описание"
// @Param altitude_km formData int true "Высота орбиты (км)"
// @Param image formData file false "Изображение"
// @Param video formData file false "Видео"
// @Success 201 {object} map[string]interface{}
// @Router /api/orbits [post]
func (h *Handler) API_AddOrbit(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	altStr := c.PostForm("altitude_km")
	altitudeKm, _ := strconv.Atoi(altStr)

	if name == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поля name и description обязательны"})
		return
	}

	var imageKey, videoKey *string

	imageFile, err := c.FormFile("image")
	if err == nil {
		ext := filepath.Ext(imageFile.Filename)
		imgName := fmt.Sprintf("img_%d%s", time.Now().Unix(), ext)

		if err := h.Repository.UploadFileToMinio(imageFile, imgName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки картинки в Minio"})
			return
		}
		imageKey = &imgName
	}

	videoFile, err := c.FormFile("video")
	if err == nil {
		ext := filepath.Ext(videoFile.Filename)
		vidName := fmt.Sprintf("vid_%d%s", time.Now().Unix(), ext)

		if err := h.Repository.UploadFileToMinio(videoFile, vidName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки видео в Minio"})
			return
		}
		videoKey = &vidName
	}

	orbit := ds.OrbitType{
		Name:        name,
		Description: description,
		AltitudeKm:  altitudeKm,
		ImageKey:    imageKey,
		VideoKey:    videoKey,
		Status:      "ACTIVE",
	}

	if err := h.Repository.DB().Create(&orbit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения в БД"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Услуга успешно добавлена",
		"data":    orbit,
	})
}

// М-М

type M2MRequest struct {
	OrbitID uint    `json:"orbit_id"`
	Payload float64 `json:"payload_tons"`
}

// API_AddMissionOrbitItem godoc
// @Summary Добавить услугу в черновик
// @Description Автоматически находит или создает черновик для текущего пользователя по JWT.
// @Security ApiKeyAuth
// @Tags Mission-M2M
// @Accept json
// @Produce json
// @Param request body M2MRequest true "Данные М-М (orbit_id, stage_order, payload)"
// @Success 201 {object} map[string]string
// @Router /api/mission-orbit-items [post]
func (h *Handler) API_AddMissionOrbitItem(c *gin.Context) {
	userIDValue, _ := c.Get("user_id")
	userID := userIDValue.(uint) 
	var req M2MRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	mission, _ := h.Repository.GetOrCreateDraftMission(userID)

	item := ds.MissionOrbitItem{
		MissionID: mission.ID, OrbitID: req.OrbitID,
	}

	h.Repository.DB().Create(&item)
	c.JSON(201, gin.H{"message": "Добавлено в заявку"})
}

// API_DeleteMissionOrbitItem godoc
// @Summary Удалить услугу из заявки
// @Description Без передачи ID заявки, ищет по активному черновику.
// @Security ApiKeyAuth
// @Tags Mission-M2M
// @Accept json
// @Produce json
// @Param request body map[string]int true "orbit_id"
// @Success 200 {object} map[string]string
// @Router /api/mission-orbit-items [delete]
func (h *Handler) API_DeleteMissionOrbitItem(c *gin.Context) {
	userIDValue, _ := c.Get("user_id")
	userID := userIDValue.(uint) 
	var req struct {
		OrbitID uint `json:"orbit_id"`
	}
	c.ShouldBindJSON(&req)

	mission, _ := h.Repository.GetOrCreateDraftMission(userID)

	h.Repository.DB().Where("mission_id = ? AND orbit_id = ?", mission.ID, req.OrbitID).Delete(&ds.MissionOrbitItem{})
	c.JSON(200, gin.H{"message": "Удалено из заявки"})
}

// API_UpdateMissionOrbitItem godoc
// @Summary Изменить этап в черновике (нагрузку или порядок)
// @Security ApiKeyAuth
// @Tags Mission-M2M
// @Accept json
// @Produce json
// @Param request body M2MRequest true "Обновленные данные"
// @Success 200 {object} map[string]string
// @Router /api/mission-orbit-items [put]
func (h *Handler) API_UpdateMissionOrbitItem(c *gin.Context) {
	userIDValue, _ := c.Get("user_id")
	userID := userIDValue.(uint) 

	var req M2MRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных: " + err.Error()})
		return
	}

	err := h.Repository.UpdateMissionItem(userID, req.OrbitID, req.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Данные этапа (нагрузка и порядок) успешно обновлены",
	})
}

// Заявки

// API_GetMissionDraft godoc
// @Summary Получить иконку корзины (Черновик)
// @Description Возвращает ID черновика и количество услуг в нём.
// @Security ApiKeyAuth
// @Tags Missions
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/mission [get]
func (h *Handler) API_GetMissionDraft(c *gin.Context) {
	userIDValue, _ := c.Get("user_id")
	userID := userIDValue.(uint) 
	mission, err := h.Repository.GetOrCreateDraftMission(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "DB error"})
		return
	}

	var count int64
	h.Repository.DB().Model(&ds.MissionOrbitItem{}).Where("mission_id = ?", mission.ID).Count(&count)

	c.JSON(200, gin.H{"draft_id": mission.ID, "items_count": count})
}

// API_GetMissionsList godoc
// @Summary Получить список заявок с фильтрацией
// @Description Клиент видит только свои, модератор - все (кроме DELETED/DRAFT).
// @Security ApiKeyAuth
// @Tags Missions
// @Produce json
// @Param status query string false "Фильтр по статусу (FORMED, COMPLETED)"
// @Param date_from query string false "Дата от (YYYY-MM-DD)"
// @Param date_to query string false "Дата до (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /api/missions [get]
func (h *Handler) API_GetMissionsList(c *gin.Context) {
	status := c.Query("status")
	dFrom := c.Query("date_from")
	dTo := c.Query("date_to")
	userIDVal, exists := c.Get("user_id")
	userRoleVal, roleExists := c.Get("user_role")

	if !exists || !roleExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не удалось определить пользователя"})
		return
	}

	userID := userIDVal.(uint)       // Приводим интерфейс к типу uint
	userRole := userRoleVal.(string)

	list, _ := h.Repository.GetMissionsList(status, dFrom, dTo, userID, userRole)
	c.JSON(200, gin.H{"data": list})
}

// API_GetMissionDetail godoc
// @Summary Получить заявку и её состав
// @Security ApiKeyAuth
// @Tags Missions
// @Produce json
// @Param id path int true "ID Заявки"
// @Success 200 {object} map[string]interface{}
// @Router /api/missions/{id} [get]
func (h *Handler) API_GetMissionDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var mission ds.Mission
	err := h.Repository.DB().
		Preload("OrbitItems.Orbit").
		First(&mission, id).Error

	if err != nil {
		c.JSON(404, gin.H{"error": "Заявка не найдена"})
		return
	}

	c.JSON(200, gin.H{"data": mission})
}

// API_UpdateMission godoc
// @Summary Обновить массу спутника в заявке
// @Security ApiKeyAuth
// @Tags Missions
// @Accept json
// @Produce json
// @Param id path int true "ID Заявки"
// @Param request body map[string]int true "satellite_mass_kg"
// @Success 200 {object} map[string]string
// @Router /api/missions/{id} [put]
func (h *Handler) API_UpdateMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		SatelliteMassKg int `json:"satellite_mass_kg"`
	}
	c.ShouldBindJSON(&req)

	h.Repository.DB().Model(&ds.Mission{}).Where("id = ?", id).Update("satellite_mass_kg", req.SatelliteMassKg)
	c.JSON(200, gin.H{"message": "Обновлено"})
}

// API_FormMission godoc
// @Summary Сформировать заявку (Переход из DRAFT в FORMED и расчеты)
// @Security ApiKeyAuth
// @Tags Missions
// @Produce json
// @Param id path int true "ID Заявки"
// @Success 200 {object} map[string]string
// @Router /api/missions/{id}/form [put]
func (h *Handler) API_FormMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Repository.FormMission(uint(id)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Сформировано и рассчитано"})
}

// API_CompleteMission godoc
// @Summary Завершить/Отклонить заявку (ТОЛЬКО МОДЕРАТОР)
// @Security ApiKeyAuth
// @Tags Missions
// @Accept json
// @Produce json
// @Param id path int true "ID Заявки"
// @Param request body map[string]string true "action: ACCEPT или REJECT"
// @Success 200 {object} map[string]string
// @Router /api/missions/{id}/complete [put]
func (h *Handler) API_CompleteMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Action string `json:"action"`
	}
	c.ShouldBindJSON(&req)

	reject := req.Action == "REJECT"

	userIDValue, _ := c.Get("user_id")
	userID := userIDValue.(uint) 
	err := h.Repository.CompleteMission(uint(id), userID, reject)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка обработана модератором"})
}

// API_DeleteMission godoc
// @Summary Удалить заявку (SQL UPDATE)
// @Security ApiKeyAuth
// @Tags Missions
// @Produce json
// @Param id path int true "ID Заявки"
// @Success 200 {object} map[string]string
// @Router /api/missions/{id} [delete]
func (h *Handler) API_DeleteMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.Repository.DB().Exec("UPDATE missions SET status = 'DELETED' WHERE id = ?", id).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Ошибка при удалении заявки"})
		return
	}

	c.JSON(200, gin.H{"message": "Заявка успешно удалена"})
}
