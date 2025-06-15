package models

import "gorm.io/gorm"

type RiskManagement struct {
	gorm.Model
	CompanyID          uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID          uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version            int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	RegulatoryChanges  string
	ComplianceStatus   string
	RegulatoryConcerns string
	SecurityAudits     string
	DataProtection     string
	SecurityIncidents  string
	KeyDependencies    string
	ContingencyPlans   string

	IsVisible  int
	IsEditable int
}
