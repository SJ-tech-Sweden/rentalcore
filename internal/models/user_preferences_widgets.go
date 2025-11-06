package models

import "time"

// DashboardWidgetPreferences stores per-user dashboard widget settings.
type DashboardWidgetPreferences struct {
	ID        uint      `json:"id" gorm:"primaryKey;column:id"`
	UserID    uint      `json:"userID" gorm:"column:user_id;uniqueIndex"`
	Widgets   []byte    `json:"widgets" gorm:"column:widgets;type:json"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName overrides the default table name for DashboardWidgetPreferences.
func (DashboardWidgetPreferences) TableName() string {
	return "user_dashboard_widgets"
}

// DashboardWidgetDefinition describes an available dashboard widget.
type DashboardWidgetDefinition struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Link        string `json:"link"`
}
