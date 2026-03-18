package repository

import (
	"errors"
	"fmt"
	"math"
	"orbit-calc/internal/app/ds"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetActiveOrbits(query string) ([]ds.OrbitType, error) {
	var orbits []ds.OrbitType
	db := r.db.Where("status = ?", "ACTIVE")

	if query != "" {
		db = db.Where("name ILIKE ?", "%"+query+"%")
	}

	err := db.Find(&orbits).Error
	return orbits, err
}

func (r *Repository) GetOrbitByID(id uint) (*ds.OrbitType, error) {
	var orbit ds.OrbitType
	err := r.db.First(&orbit, id).Error
	if err != nil {
		return nil, err
	}
	return &orbit, nil
}

func (r *Repository) GetOrCreateDraftMission(userID uint) (*ds.Mission, error) {
	var mission ds.Mission

	err := r.db.Where("author_id = ? AND status = ?", userID, "DRAFT").First(&mission).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newMission := ds.Mission{
			AuthorID:  userID,
			Status:    "DRAFT",
			CreatedAt: time.Now(),
		}
		if err = r.db.Create(&newMission).Error; err != nil {
			return nil, err
		}
		return &newMission, nil
	}

	return &mission, err
}

func (r *Repository) AddOrbitToMission(missionID, orbitID uint, payload string) error {
	var orbit ds.OrbitType
	r.db.First(&orbit, orbitID)

	const EarthRadius = 6371.0
	const mu = 398600.44  // mu = G * Mз
	distanceFromCentre := EarthRadius + float64(orbit.AltitudeKm)

	velocity := math.Sqrt(mu / distanceFromCentre)
	period := (2 * math.Pi * distanceFromCentre * math.Sqrt(distanceFromCentre / mu)) / 60.0  // 2piR / v + перевод в минуты

	item := ds.MissionOrbitItem{
		MissionID: missionID,
		OrbitID:   orbitID,
		Payload:   payload,
		Velocity:  velocity,
		Period:    period,
	}

	return r.db.Create(&item).Error
}

func (r *Repository) GetMissionByID(missionID uint) (*ds.Mission, error) {
	var mission ds.Mission
	err := r.db.Preload("OrbitItems.Orbit").Where("status != 'DELETED'").First(&mission, missionID).Error
	if err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *Repository) GetDraftMissionWithItems(userID uint) (*ds.Mission, error) {
	var mission ds.Mission

	err := r.db.Preload("OrbitItems.Orbit").Where("author_id = ? AND status = ?", userID, "DRAFT").First(&mission).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &mission, err
}

func (r *Repository) DeleteMission(missionID uint) error {
	result := r.db.Exec("UPDATE missions SET status = ? WHERE id = ?", "DELETED", missionID)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка с ID %d не найдена", missionID)
	}

	return nil
}
