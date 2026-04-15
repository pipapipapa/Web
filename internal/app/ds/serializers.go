package ds

type MissionListDTO struct {
	ID                   uint    `json:"id"`
	Status               string  `json:"status"`
	FormedAt             string  `json:"formed_at,omitempty"`
	AuthorLogin          string  `json:"author_login"`
	ModeratorLogin       string  `json:"moderator_login,omitempty"`
	SatelliteName        string  `json:"satellite_name"`
	
	CalculatedItemsCount int     `json:"calculated_items_count"` 
}