package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const MinioBaseURL = "http://127.0.0.1:9000/orbits/"


func (h *Handler) ShowOrbitsPage(c *gin.Context) {
	const currentUserID = 1 
	
	query := c.Query("q")
	orbits, err := h.Repository.GetActiveOrbits(query)
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка получения данных")
		log.Println(err)
		return
	}

	draft, _ := h.Repository.GetDraftMissionWithItems(currentUserID)
	
	var missionCount int
	var draftID uint
	if draft != nil {
		missionCount = len(draft.OrbitItems)
		draftID = draft.ID
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Orbits":       orbits,
		"Query":        query,
		"MissionCount": missionCount,
		"MinioURL":     MinioBaseURL,
		"DraftID":      draftID,
	})
}

func (h *Handler) ShowOrbitDetailPage(c *gin.Context) {
	const currentUserID = 1 
	id, _ := strconv.Atoi(c.Param("id"))
	orbit, err := h.Repository.GetOrbitByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Услуга не найдена")
		return
	}
	
	draft, _ := h.Repository.GetDraftMissionWithItems(currentUserID)

	var missionCount int
	var draftID uint
	if draft != nil {
		missionCount = len(draft.OrbitItems)
		draftID = draft.ID
	}

	c.HTML(http.StatusOK, "orbit.html", gin.H{
		"Orbit":        orbit,
		"MissionCount": missionCount,
		"MinioURL":     MinioBaseURL,
		"DraftID":      draftID,
	})
}

func (h *Handler) ShowMissionPage(c *gin.Context) {
	const currentUserID = 1 
	
	idParam := c.Param("id")
	missionID, _ := strconv.Atoi(idParam)

	mission, err := h.Repository.GetMissionByID(uint(missionID))
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	draft, _ := h.Repository.GetDraftMissionWithItems(currentUserID)
	var missionCount int
	if draft != nil {
		missionCount = len(draft.OrbitItems)
	}

	c.HTML(http.StatusOK, "mission.html", gin.H{
		"Mission":      mission,
		"MissionCount": missionCount,
		"MinioURL":     MinioBaseURL,
	})
}

func (h *Handler) AddOrbitToMission(c *gin.Context) {
	const currentUserID = 1 
	orbitID, _ := strconv.Atoi(c.PostForm("orbit_id"))
	
	mission, err := h.Repository.GetOrCreateDraftMission(uint(currentUserID))
	if err != nil {
		c.String(500, "Ошибка при подготовке миссии")
		return
	}

	err = h.Repository.AddOrbitToMission(mission.ID, uint(orbitID))
	if err != nil {
		log.Println("Предупреждение:", err)
	}

	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) DeleteMission(c *gin.Context) {
	missionID, _ := strconv.Atoi(c.PostForm("mission_id"))
	
	err := h.Repository.DeleteMission(uint(missionID))
	if err != nil {
		c.String(http.StatusInternalServerError, "Не удалось удалить заявку")
		log.Println(err)
		return
	}
	
	c.Redirect(http.StatusFound, "/")
}