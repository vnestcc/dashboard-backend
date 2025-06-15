package models

import "gorm.io/gorm"

type ProductDevelopment struct {
	gorm.Model
	CompanyID           uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID           uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version             int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	MilestonesAchieved  string
	MilestonesMissed    string
	Roadmap             string
	ActiveUsers         string
	EngagementMetrics   string
	NPS                 string
	FeatureAdoption     string
	TechnicalChallenges string
	TechnicalDebt       string
	ProductBottlenecks  string

	IsVisible  int
	IsEditable int
}
