package models

import "gorm.io/gorm"

type UnitEconomics struct {
	gorm.Model
	CompanyID  uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID  uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version    int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	CAC        string
	CACChange  string
	LTV        string
	LTVRatio   string
	CACPayback string
	ARPU       string

	MarketingBreakdowns []MarketingBreakdown

	IsVisible  int
	IsEditable int
}

type MarketingBreakdown struct {
	gorm.Model
	UnitEconomicsID uint
	Channel         string
	Spend           string
	Budget          string
	CAC             string
}
