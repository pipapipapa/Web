package handler

import (
	"net/http"
	"strings"

	"orbit-calc/internal/app/repository"

	"github.com/gin-gonic/gin"
)

var orbits = []repository.OrbitType{
	{ID: 1, Name: "Низкая Околоземная До 250 кг", AltitudeKm: 400, Description: "Вывод миниспутника и меньше. Орбита для МКС и Starlink.", ImageKey: "leo1.jpg", VideoKey: "leo.mp4", Type: "loe", Weight: "10-500 кг", Velocity: 7.66, Period: 92.5},
	{ID: 2, Name: "Геостационарная До 250 кг", AltitudeKm: 35800, Description: "Вывод миниспутника и меньше. Спутник висит над одной точкой.", ImageKey: "geo1.jpg", VideoKey: "geo.mp4", Type: "geo", Weight: "10-500 кг", Velocity: 17.06, Period: 122.5},
	{ID: 3, Name: "Солнце-синхронная До 500 кг", AltitudeKm: 800, Description: "Вывод миниспутника и меньше. Пролетает над точкой в одно время.", ImageKey: "sso1.jpg", VideoKey: "sso.mp4", Type: "sso", Weight: "250-500 кг", Velocity: 56.9, Period: 0},
	{ID: 4, Name: "Низкая Околоземная До 1000 кг", AltitudeKm: 400, Description: "Средний спутник. Спутник висит над одной точкой.", ImageKey: "leo2.jpg", VideoKey: "leo.mp4", Type: "loe", Weight: "500-1000 кг", Velocity: 121.8, Period: 99},
	{ID: 5, Name: "Геостационарная До 1000 кг", AltitudeKm: 35800, Description: "Средний спутник. Орбита для МКС и Starlink.", ImageKey: "geo2.jpg", VideoKey: "geo.mp4", Type: "geo", Weight: "500-1000 кг", Velocity: 29.12, Period: 43},
	{ID: 6, Name: "Геостационарная До 2500 кг", AltitudeKm: 35800, Description: "Большой спутник. Спутник висит над одной точкой.", ImageKey: "geo3.jpg", VideoKey: "geo.mp4", Type: "geo", Weight: "1000-2500 кг", Velocity: 30.3, Period: 87},
	{ID: 7, Name: "Низкая Околоземная До 3000 кг", AltitudeKm: 400, Description: "Тяжелые спутники. Орбита для МКС и Starlink.", ImageKey: "leo3.jpg", VideoKey: "leo.mp4", Type: "loe", Weight: "1500-3000 кг", Velocity: 5.8, Period: 27},
}

var currentOrbitMission = repository.OrbitMissionProject{
	ID:                 101,
	SatelliteName:      "Sat-X Experimental",
	SatelliteMassKg:    500,
	CalculatedAltitude: 35786,
	CalculatedVelocity: 7.44,
	CalculatedPeriod:   82.6,
	SelectedOrbits: []repository.OrbitMissionItem{
		{
			Orbit:				orbits[0],
			StageOrder:			1,
			PayloadDescription: "Вывод ракетой-носителем, развертывание панелей",
			DeltaV:		9.40,
		},
		{
			Orbit:				orbits[1],
			StageOrder:			2,
			PayloadDescription: "Активация двигателя, переход на ГСО",
			DeltaV:		4.30,
		},
	},
}

type Handler struct{}

func (h *Handler) GetOrbits(c *gin.Context) {
	query := c.Query("q")
	var filtered []repository.OrbitType

	if query != "" {
		for _, o := range orbits {
			if strings.Contains(strings.ToLower(o.Name), strings.ToLower(query)) {
				filtered = append(filtered, o)
			}
		}
	} else {
		filtered = orbits
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Orbits":       filtered,
		"MissionCount": len(currentOrbitMission.SelectedOrbits),
		"Query":        query,
		"MinioURL":     repository.MinioBaseURL, 
	})
}

func (h *Handler) GetOrbitDetail(c *gin.Context) {
	id := c.Param("id")
	var orbit repository.OrbitType
	for _, o := range orbits {
		if string(rune(o.ID+'0')) == id || id == "1" && o.ID == 1 || id == "2" && o.ID == 2 {
			orbit = o
			break
		}
	}

	c.HTML(http.StatusOK, "detail.html", gin.H{
		"Orbit":    orbit,
		"MinioURL": repository.MinioBaseURL,
		"MissionCount": len(currentOrbitMission.SelectedOrbits),
	})
}

func (h *Handler) GetOrbitMission(c *gin.Context) {
	c.HTML(http.StatusOK, "mission.html", gin.H{
		"Mission":  currentOrbitMission,
		"MinioURL": repository.MinioBaseURL,
		"MissionCount": len(currentOrbitMission.SelectedOrbits),
	})
}
