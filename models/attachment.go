package models

import "gorm.io/gorm"

type Attachment struct {
	gorm.Model
	CompanyID            uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID            uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version              int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	FinancialStatements  string
	PitchDeck            string
	ProductRoadmap       string
	PerformanceDashboard string
	OrgChart             string

	IsVisible  int
	IsEditable int
}
