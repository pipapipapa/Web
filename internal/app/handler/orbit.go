package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const MinioBaseURL = "http://127.0.0.1:9000/orbits/"

// --- 3 GET-метода ---

// ShowOrbitsPage - отображает главную страницу со списком услуг
func (h *Handler) ShowOrbitsPage(c *gin.Context) {
	// Временно хардкодим ID пользователя, пока нет авторизации
	const currentUserID = 1 
	
	query := c.Query("q")
	orbits, err := h.Repository.GetActiveOrbits(query)
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка получения данных")
		log.Println(err)
		return
	}

	// Получаем текущую заявку для отображения в "корзине"
	draft, _ := h.Repository.GetUserDraft(currentUserID)
	
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

// ShowOrbitDetailPage - страница с детальной информацией
func (h *Handler) ShowOrbitDetailPage(c *gin.Context) {
	const currentUserID = 1 
	id, _ := strconv.Atoi(c.Param("id"))
	orbit, err := h.Repository.GetOrbitByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Услуга не найдена")
		return
	}
	
	draft, _ := h.Repository.GetUserDraft(currentUserID)

	var missionCount int
	var draftID uint
	if draft != nil {
		missionCount = len(draft.OrbitItems)
		draftID = draft.ID
	}

	c.HTML(http.StatusOK, "detail.html", gin.H{
		"Orbit":        orbit,
		"MissionCount": missionCount,
		"MinioURL":     MinioBaseURL,
		"DraftID":      draftID,
	})
}

func (h *Handler) ShowMissionPage(c *gin.Context) {
	// Берем ID прямо из URL (/mission/105)
	idParam := c.Param("id")
	missionID, _ := strconv.Atoi(idParam)

	mission, err := h.Repository.GetMissionByID(uint(missionID))
	if err != nil {
		// Если заявка DELETED или не существует — перекидываем на главную
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Для шапки на этой странице тоже нужно знать, есть ли у юзера другой активный черновик
	draft, _ := h.Repository.GetUserDraft(1)
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
// --- 2 POST-метода ---

// AddOrbitToMission - добавляет услугу в заявку
func (h *Handler) AddOrbitToMission(c *gin.Context) {
	const currentUserID = 1 
	orbitID, _ := strconv.Atoi(c.PostForm("orbit_id"))
	
	// Вся логика теперь в двух строчках репозитория:
	
	// 1. Дай мне черновик (создай если нет)
	mission, err := h.Repository.GetOrCreateDraftMission(uint(currentUserID))
	if err != nil {
		c.String(500, "Ошибка при подготовке миссии")
		return
	}

	// 2. Добавь орбиту в план (там же произойдут расчеты)
	err = h.Repository.AddOrbitToMission(mission.ID, uint(orbitID), "Стандартная научная нагрузка")
	if err != nil {
		// Ошибка может быть, если такая орбита уже в плане (составной ключ)
		log.Println("Предупреждение:", err)
	}

	c.Redirect(http.StatusFound, "/mission/"+strconv.Itoa(int(mission.ID)))
}

// DeleteMission - логически удаляет заявку
func (h *Handler) DeleteMission(c *gin.Context) {
	missionID, _ := strconv.Atoi(c.PostForm("mission_id"))
	
	err := h.Repository.DeleteMission(uint(missionID))
	if err != nil {
		c.String(http.StatusInternalServerError, "Не удалось удалить заявку")
		log.Println(err)
		return
	}
	
	// Перенаправляем на главную, где удаленной заявки уже не будет видно
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) RemoveItem(c *gin.Context) {
	missionID, _ := strconv.Atoi(c.PostForm("mission_id"))
	orbitID, _ := strconv.Atoi(c.PostForm("orbit_id"))

	h.Repository.RemoveItemFromMission(uint(missionID), uint(orbitID))
	
	c.Redirect(302, "/mission/"+strconv.Itoa(missionID))
}

// Сформировать (Ввод массы)
func (h *Handler) FormMission(c *gin.Context) {
	missionID, _ := strconv.Atoi(c.PostForm("mission_id"))
	mass, _ := strconv.Atoi(c.PostForm("mass"))

	// Собираем мапу Payload из всех полей payload_{id}
	payloads := make(map[uint]string)
	for key, values := range c.Request.PostForm {
		if strings.HasPrefix(key, "payload_") {
			id, _ := strconv.Atoi(strings.TrimPrefix(key, "payload_"))
			payloads[uint(id)] = values[0]
		}
	}

	h.Repository.FormMission(uint(missionID), mass, payloads)
	c.Redirect(302, "/mission/"+strconv.Itoa(missionID))
}

// Завершить (Расчет)
func (h *Handler) CompleteMission(c *gin.Context) {
	missionID, _ := strconv.Atoi(c.PostForm("mission_id"))

	h.Repository.CalculateAndComplete(uint(missionID))
	c.Redirect(302, "/mission/"+strconv.Itoa(missionID))
}
