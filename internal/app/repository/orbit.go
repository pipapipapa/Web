package repository

import (
	"errors"
	"fmt"
	"orbit-calc/internal/app/ds"
	"time"
	"math"

	"gorm.io/gorm"
)

// --- Методы для Услуг (Орбит) ---

// GetActiveOrbits - получает все активные орбиты с возможностью поиска (использует ORM)
func (r *Repository) GetActiveOrbits(query string) ([]ds.OrbitType, error) {
	var orbits []ds.OrbitType
	db := r.db.Where("status = ?", "ACTIVE")

	if query != "" {
		db = db.Where("name ILIKE ?", "%"+query+"%")
	}

	err := db.Find(&orbits).Error
	return orbits, err
}

// GetOrbitByID - получает одну орбиту по ID (использует ORM)
func (r *Repository) GetOrbitByID(id uint) (*ds.OrbitType, error) {
	var orbit ds.OrbitType
	err := r.db.First(&orbit, id).Error
	if err != nil {
		return nil, err
	}
	return &orbit, nil
}


// --- Методы для Заявок (Миссий) и М-М ---

// GetOrCreateDraftMission - находит черновик для пользователя или создает новый (использует ORM)
func (r *Repository) GetOrCreateDraftMission(userID uint) (*ds.Mission, error) {
	var mission ds.Mission
	
	// Ищем черновик для пользователя
	err := r.db.Where("author_id = ? AND status = ?", userID, "DRAFT").First(&mission).Error
	
	// Если черновик не найден, создаем его
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
	// 1. Получаем высоту орбиты для расчетов
	var orbit ds.OrbitType
	r.db.First(&orbit, orbitID)

	// 2. Расчеты (Формулы Кеплера)
	const R = 6371.0
	const mu = 398600.44
	a := R + float64(orbit.AltitudeKm)
	
	velocity := math.Sqrt(mu / a)
	period := (2 * math.Pi * math.Sqrt(math.Pow(a, 3)/mu)) / 60.0

	// 3. Создание записи М-М
	item := ds.MissionOrbitItem{
		MissionID: missionID,
		OrbitID:   orbitID,
		Payload:   payload,
		Velocity:  velocity,
		Period:    period,
	}

	return r.db.Create(&item).Error
}

// GetMissionByID - получает миссию со всем составом для страницы корзины
func (r *Repository) GetMissionByID(missionID uint) (*ds.Mission, error) {
	var mission ds.Mission
	// Загружаем саму миссию + связанные OrbitItems + сами Орбиты внутри них
	err := r.db.Preload("OrbitItems.Orbit").Where("status != 'DELETED'").First(&mission, missionID).Error
	if err != nil {
		return nil, err
	}
	return &mission, nil
}


// GetDraftMissionWithItems - получает черновик со всеми вложенными услугами (использует ORM Preload)
func (r *Repository) GetDraftMissionWithItems(userID uint) (*ds.Mission, error) {
	var mission ds.Mission
	
	// `Preload("OrbitItems.Orbit")` — магия GORM, которая подгружает данные из связанных таблиц
	err := r.db.Preload("OrbitItems.Orbit").Where("author_id = ? AND status = ?", userID, "DRAFT").First(&mission).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Если черновика нет, это не ошибка
	}
	
	return &mission, err
}

// DeleteMissionLogically - логическое удаление заявки (использует чистый SQL UPDATE)
func (r *Repository) DeleteMission(missionID uint) error {
	// Выполняем SQL-запрос напрямую, как требуется в задании
	result := r.db.Exec("UPDATE missions SET status = ? WHERE id = ?", "DELETED", missionID)
	
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка с ID %d не найдена", missionID)
	}
	
	return nil
}

// Удаление одной услуги из миссии (ORM)
func (r *Repository) RemoveItemFromMission(missionID, orbitID uint) error {
	return r.db.Where("mission_id = ? AND orbit_id = ?", missionID, orbitID).Delete(&ds.MissionOrbitItem{}).Error
}

// FormMission - Установка массы и полезной нагрузки для каждого этапа
func (r *Repository) FormMission(missionID uint, mass int, payloads map[uint]string) error {
	// 1. Обновляем общие данные миссии (Заявка)
	err := r.db.Model(&ds.Mission{}).Where("id = ?", missionID).
		Updates(map[string]interface{}{
			"satellite_mass_kg": mass,
			"status":            "FORMED",
			"formed_at":         time.Now(),
		}).Error
	if err != nil {
		return err
	}

	// 2. Обновляем полезную нагрузку для каждого этапа (М-М)
	for orbitID, text := range payloads {
		err := r.db.Model(&ds.MissionOrbitItem{}).
			Where("mission_id = ? AND orbit_id = ?", missionID, orbitID).
			Update("payload", text).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) GetUserDraft(userID uint) (*ds.Mission, error) {
	var mission ds.Mission
	err := r.db.Preload("OrbitItems").
		Where("author_id = ? AND status = 'DRAFT'", userID).
		First(&mission).Error
	if err != nil {
		return nil, err // Если не нашли DRAFT — это не ошибка, просто вернем nil
	}
	return &mission, nil
}

// CalculateAndComplete - Финальный расчет по формулам
func (r *Repository) CalculateAndComplete(missionID uint) error {
	var items []ds.MissionOrbitItem
	// Достаем все этапы миссии вместе с данными об орбитах
	r.db.Preload("Orbit").Where("mission_id = ?", missionID).Find(&items)

	const R = 6371.0
	const mu = 398600.44

	for _, item := range items {
		a := R + float64(item.Orbit.AltitudeKm)
		v := math.Sqrt(mu / a)
		t := (2 * math.Pi * math.Sqrt(math.Pow(a, 3)/mu)) / 60.0

		// Обновляем результаты в М-М
		r.db.Model(&item).Where("mission_id = ? AND orbit_id = ?", item.MissionID, item.OrbitID).
			Updates(map[string]interface{}{
				"velocity": v,
				"period":   t,
			})
	}

	// Переводиm миссию в статус COMPLETED
	return r.db.Model(&ds.Mission{}).Where("id = ?", missionID).
		Updates(map[string]interface{}{
			"status":       "COMPLETED",
			"completed_at": time.Now(),
		}).Error
}