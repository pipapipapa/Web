package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"orbit-calc/internal/app/auth"
	"orbit-calc/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// Услуги

func (h *Handler) API_GetOrbits(c *gin.Context) {
	q := c.Query("search")
	orbits, _ := h.Repository.GetOrbits(q)
	c.JSON(200, gin.H{"data": orbits})
}

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

func (h *Handler) API_AddMissionOrbitItem(c *gin.Context) {
	userID := auth.GetCurrentUserID()
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

func (h *Handler) API_DeleteMissionOrbitItem(c *gin.Context) {
	userID := auth.GetCurrentUserID()
	var req struct {
		OrbitID uint `json:"orbit_id"`
	}
	c.ShouldBindJSON(&req)

	mission, _ := h.Repository.GetOrCreateDraftMission(userID)

	h.Repository.DB().Where("mission_id = ? AND orbit_id = ?", mission.ID, req.OrbitID).Delete(&ds.MissionOrbitItem{})
	c.JSON(200, gin.H{"message": "Удалено из заявки"})
}

func (h *Handler) API_UpdateMissionOrbitItem(c *gin.Context) {
	userID := auth.GetCurrentUserID()

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

func (h *Handler) API_GetMissionDraft(c *gin.Context) {
	userID := auth.GetCurrentUserID()
	mission, err := h.Repository.GetOrCreateDraftMission(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "DB error"})
		return
	}

	var count int64
	h.Repository.DB().Model(&ds.MissionOrbitItem{}).Where("mission_id = ?", mission.ID).Count(&count)

	c.JSON(200, gin.H{"draft_id": mission.ID, "items_count": count})
}

func (h *Handler) API_GetMissionsList(c *gin.Context) {
	status := c.Query("status")
	dFrom := c.Query("date_from")
	dTo := c.Query("date_to")

	list, _ := h.Repository.GetMissionsList(status, dFrom, dTo)
	c.JSON(200, gin.H{"data": list})
}

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

func (h *Handler) API_UpdateMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		SatelliteMassKg int `json:"satellite_mass_kg"`
	}
	c.ShouldBindJSON(&req)

	h.Repository.DB().Model(&ds.Mission{}).Where("id = ?", id).Update("satellite_mass_kg", req.SatelliteMassKg)
	c.JSON(200, gin.H{"message": "Обновлено"})
}

func (h *Handler) API_FormMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Repository.FormMission(uint(id)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Сформировано и рассчитано"})
}

func (h *Handler) API_CompleteMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Action string `json:"action"`
	}
	c.ShouldBindJSON(&req)

	reject := req.Action == "REJECT"

	err := h.Repository.CompleteMission(uint(id), auth.GetCurrentModeratorID(), reject)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка обработана модератором"})
}

func (h *Handler) API_DeleteMission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.Repository.DB().Exec("UPDATE missions SET status = 'DELETED' WHERE id = ?", id).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Ошибка при удалении заявки"})
		return
	}

	c.JSON(200, gin.H{"message": "Заявка успешно удалена"})
}
