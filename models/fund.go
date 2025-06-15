package models

import "gorm.io/gorm"

type FundraisingStatus struct {
	gorm.Model
	CompanyID             uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID             uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version               int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	LastRound             string
	CurrentInvestors      string
	InvestorRelations     string
	NextRound             string
	TargetAmount          string
	InvestorPipeline      string
	ValuationExpectations string

	IsVisible  int
	IsEditable int
}
