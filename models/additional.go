package models

import "gorm.io/gorm"

type AdditionalInfo struct {
	gorm.Model
	CompanyID                uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID                uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version                  int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	GrowthChallenges         string
	SupportNeeded            string
	PolicyChanges            string
	PolicyImpact             string
	MitigationStrategies     string
	NewInitiatives           string
	InitiativeProgress       string
	BusinessModelAdjustments string

	IsVisible  int
	IsEditable int
}
