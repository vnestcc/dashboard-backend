package models

import (
	"fmt"

	"gorm.io/gorm"
)

type FinancialHealth struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	CashBalance           string `json:"cash_balance"`
	BurnRate              string `json:"burn_rate"`
	CashRunway            string `json:"cash_runway"`
	BurnRateChange        string `json:"burn_rate_change"`
	QuarterlyRevenue      string `json:"quarterly_revenue"`
	RevenueGrowth         string `json:"revenue_growth"`
	GrossMargin           string `json:"gross_margin"`
	NetMargin             string `json:"net_margin"`
	ProfitabilityTimeline string `json:"profitability_timeline"`

	RevenueBreakdowns []RevenueBreakdown `json:"revenue_breakdowns"`

	IsVisible  uint16 `gorm:"default:1023" json:"-"`
	IsEditable uint16 `gorm:"default:1023" json:"-"`
}

type RevenueBreakdown struct {
	gorm.Model
	FinancialHealthID uint   `json:"financial_health_id"`
	Product           string `json:"product"`
	Revenue           string `json:"revenue"`
	Percentage        string `json:"percentage"`
}

func (f *FinancialHealth) TableName() string {
	return "finance"
}

func (f *FinancialHealth) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"CashBalance",
		"BurnRate",
		"CashRunway",
		"BurnRateChange",
		"QuarterlyRevenue",
		"RevenueGrowth",
		"GrossMargin",
		"NetMargin",
		"ProfitabilityTimeline",
		"RevenueBreakdowns",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if f.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (f *FinancialHealth) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"CashBalance":           f.CashBalance,
			"BurnRate":              f.BurnRate,
			"CashRunway":            f.CashRunway,
			"BurnRateChange":        f.BurnRateChange,
			"QuarterlyRevenue":      f.QuarterlyRevenue,
			"RevenueGrowth":         f.RevenueGrowth,
			"GrossMargin":           f.GrossMargin,
			"NetMargin":             f.NetMargin,
			"ProfitabilityTimeline": f.ProfitabilityTimeline,
			"RevenueBreakdowns":     f.RevenueBreakdowns,
		}
	}

	result := make(map[string]any)
	if f.IsVisible&(1<<0) != 0 {
		result["CashBalance"] = f.CashBalance
	}
	if f.IsVisible&(1<<1) != 0 {
		result["BurnRate"] = f.BurnRate
	}
	if f.IsVisible&(1<<2) != 0 {
		result["CashRunway"] = f.CashRunway
	}
	if f.IsVisible&(1<<3) != 0 {
		result["BurnRateChange"] = f.BurnRateChange
	}
	if f.IsVisible&(1<<4) != 0 {
		result["QuarterlyRevenue"] = f.QuarterlyRevenue
	}
	if f.IsVisible&(1<<5) != 0 {
		result["RevenueGrowth"] = f.RevenueGrowth
	}
	if f.IsVisible&(1<<6) != 0 {
		result["GrossMargin"] = f.GrossMargin
	}
	if f.IsVisible&(1<<7) != 0 {
		result["NetMargin"] = f.NetMargin
	}
	if f.IsVisible&(1<<8) != 0 {
		result["ProfitabilityTimeline"] = f.ProfitabilityTimeline
	}
	if f.IsVisible&(1<<9) != 0 {
		result["RevenueBreakdowns"] = f.RevenueBreakdowns
	}
	return result
}

func (f *FinancialHealth) EditableFilter() error {
	var errFields []string

	if f.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "CashBalance")
	}
	if f.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "BurnRate")
	}
	if f.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "CashRunway")
	}
	if f.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "BurnRateChange")
	}
	if f.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "QuarterlyRevenue")
	}
	if f.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "RevenueGrowth")
	}
	if f.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "GrossMargin")
	}
	if f.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "NetMargin")
	}
	if f.IsEditable&(1<<8) == 0 {
		errFields = append(errFields, "ProfitabilityTimeline")
	}
	if f.IsEditable&(1<<9) == 0 {
		errFields = append(errFields, "RevenueBreakdowns")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
