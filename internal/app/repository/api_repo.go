package repository

import (
	"errors"
	"fmt"
	"math"
	"orbit-calc/internal/app/ds"
	"time"
)


func (r *Repository) GetOrbits(nameFilter string) ([]ds.OrbitType, error) {
	var orbits[]ds.OrbitType
	q := r.db.Where("status != 'DELETED'")
	if nameFilter != "" {
		q = q.Where("name ILIKE ?", "%"+nameFilter+"%")
	}
	err := q.Find(&orbits).Error
	return orbits, err
}

func (r *Repository) GetMissionsList(status, dateFrom, dateTo string, userID uint, userRole string) ([]ds.MissionListDTO, error) {
	var missions[]ds.Mission
	q := r.db.Preload("Author").Preload("Moderator").Preload("OrbitItems").
		Where("status NOT IN ('DELETED', 'DRAFT')")

	if userRole == "CLIENT" {
		q = q.Where("author_id = ?", userID)
	}

	if status != "" {
		q = q.Where("status = ?", status)
	}
	if dateFrom != "" && dateTo != "" {
		q = q.Where("formed_at BETWEEN ? AND ?", dateFrom, dateTo)
	}

	if err := q.Find(&missions).Error; err != nil {
		return nil, err
	}

	var response[]ds.MissionListDTO
	for _, m := range missions {
		dto := ds.MissionListDTO{
			ID:            m.ID,
			Status:        m.Status,
			AuthorLogin:   m.Author.Login,
		}
		if m.FormedAt != nil { dto.FormedAt = m.FormedAt.Format(time.RFC3339) }
		if m.Moderator != nil { dto.ModeratorLogin = m.Moderator.Login }

		calcCount := 0
		for _, item := range m.OrbitItems {
			if item.Velocity != nil {
				calcCount++
			}
		}
		dto.CalculatedItemsCount = calcCount

		response = append(response, dto)
	}
	return response, nil
}

func (r *Repository) FormMission(missionID uint) error {
	var mission ds.Mission
	if err := r.db.Preload("OrbitItems.Orbit").First(&mission, missionID).Error; err != nil {
		return err
	}

	if len(mission.OrbitItems) == 0 || mission.SatelliteMassKg == nil || *mission.SatelliteMassKg <= 0 {
		return errors.New("misson lacks items or satellite mass")
	}

	const R, mu = 6371.0, 398600.44
	for _, item := range mission.OrbitItems {
		a := R + float64(item.Orbit.AltitudeKm)
		v := math.Sqrt(mu / a)
		p := (2 * math.Pi * math.Sqrt(math.Pow(a, 3)/mu)) / 60.0

		r.db.Model(&ds.MissionOrbitItem{}).
			Where("mission_id = ? AND orbit_id = ?", missionID, item.OrbitID).
			Updates(map[string]interface{}{"velocity": v, "period": p})
	}

	now := time.Now()
	return r.db.Model(&mission).Updates(map[string]interface{}{
		"status":    "FORMED",
		"formed_at": now,
	}).Error
}

func (r *Repository) CompleteMission(missionID, moderatorID uint, reject bool) error {
	var mission ds.Mission
	if err := r.db.First(&mission, missionID).Error; err != nil {
		return errors.New("заявка не найдена")
	}

	if mission.Status != "FORMED" {
		return fmt.Errorf("нельзя завершить заявку в статусе %s. Ожидается статус FORMED", mission.Status)
	}

	status := "COMPLETED"
	if reject {
		status = "REJECTED"
	}
	
	now := time.Now()

	return r.db.Model(&mission).Updates(map[string]interface{}{
		"status":       status,
		"completed_at": now,
		"moderator_id": moderatorID,
	}).Error
}

func (r *Repository) UpdateMissionItem(userID, orbitID uint, payload float64) error {
	var mission ds.Mission
	err := r.db.Where("author_id = ? AND status = 'DRAFT'", userID).First(&mission).Error
	if err != nil {
		return errors.New("активный черновик не найден")
	}

	result := r.db.Model(&ds.MissionOrbitItem{}).
		Where("mission_id = ? AND orbit_id = ?", mission.ID, orbitID).
		Updates(map[string]interface{}{
			"payload":     payload,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("указанная орбита не найдена в вашем черновике")
	}

	return nil
}