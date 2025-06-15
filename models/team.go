package models

import "gorm.io/gorm"

type TeamPerformance struct {
	gorm.Model
	CompanyID              uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID              uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version                int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	TeamSize               string
	NewHires               string
	Turnover               string
	VacantPositions        string
	LeadershipAlignment    string
	TeamStrengths          string
	SkillGaps              string
	DevelopmentInitiatives string

	IsVisible  int
	IsEditable int
}
