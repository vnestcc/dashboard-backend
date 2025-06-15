package models

import "gorm.io/gorm"

type FinancialHealth struct {
	gorm.Model
	CompanyID             uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID             uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version               int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	CashBalance           string
	BurnRate              string
	CashRunway            string
	BurnRateChange        string
	QuarterlyRevenue      string
	RevenueGrowth         string
	GrossMargin           string
	NetMargin             string
	ProfitabilityTimeline string

	RevenueBreakdowns []RevenueBreakdown

	IsVisible  int
	IsEditable int
}

type RevenueBreakdown struct {
	gorm.Model
	FinancialHealthID uint
	Product           string
	Revenue           string
	Percentage        string
}
