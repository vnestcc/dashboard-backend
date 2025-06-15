package models

import "gorm.io/gorm"

type CompetitiveLandscape struct {
	gorm.Model
	CompanyID uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version   int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	NewCompetitors       string
	CompetitorStrategies string
	MarketShifts         string
	Differentiators      string
	Threats              string
	DefensiveStrategies  string

	IsVisible  int
	IsEditable int
}
