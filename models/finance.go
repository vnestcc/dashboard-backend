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
	gorm.Model        `json:"-"`
	FinancialHealthID uint   `json:"-"`
	Product           string `json:"product"`
	Revenue           string `json:"revenue"`
	Percentage        string `json:"percentage"`
}

func (f *FinancialHealth) TableName() string {
	return "finance"
}

func (f *FinancialHealth) EditableList() []string {
	fields := []string{
		"cash_balance",
		"burn_rate",
		"cash_runway",
		"burn_rate_change",
		"quarterly_revenue",
		"revenue_growth",
		"gross_margin",
		"net_margin",
		"profitability_timeline",
		"revenue_breakdowns",
	}
	var visibleFields []string
	for i, field := range fields {
		if f.IsEditable&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (f *FinancialHealth) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"cash_balance",
		"burn_rate",
		"cash_runway",
		"burn_rate_change",
		"quarterly_revenue",
		"revenue_growth",
		"gross_margin",
		"net_margin",
		"profitability_timeline",
		"revenue_breakdowns",
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
			"version":                f.Version,
			"cash_balance":           f.CashBalance,
			"burn_rate":              f.BurnRate,
			"cash_runway":            f.CashRunway,
			"burn_rate_change":       f.BurnRateChange,
			"quarterly_revenue":      f.QuarterlyRevenue,
			"revenue_growth":         f.RevenueGrowth,
			"gross_margin":           f.GrossMargin,
			"net_margin":             f.NetMargin,
			"profitability_timeline": f.ProfitabilityTimeline,
			"revenue_breakdowns":     f.RevenueBreakdowns,
		}
	}

	result := make(map[string]any)
	result["version"] = f.Version
	if f.IsVisible&(1<<0) != 0 {
		result["cash_balance"] = f.CashBalance
	}
	if f.IsVisible&(1<<1) != 0 {
		result["burn_rate"] = f.BurnRate
	}
	if f.IsVisible&(1<<2) != 0 {
		result["cash_runway"] = f.CashRunway
	}
	if f.IsVisible&(1<<3) != 0 {
		result["burn_rate_change"] = f.BurnRateChange
	}
	if f.IsVisible&(1<<4) != 0 {
		result["quarterly_revenue"] = f.QuarterlyRevenue
	}
	if f.IsVisible&(1<<5) != 0 {
		result["revenue_growth"] = f.RevenueGrowth
	}
	if f.IsVisible&(1<<6) != 0 {
		result["gross_margin"] = f.GrossMargin
	}
	if f.IsVisible&(1<<7) != 0 {
		result["net_margin"] = f.NetMargin
	}
	if f.IsVisible&(1<<8) != 0 {
		result["profitability_timeline"] = f.ProfitabilityTimeline
	}
	if f.IsVisible&(1<<9) != 0 {
		result["revenue_breakdowns"] = f.RevenueBreakdowns
	}
	return result
}

func (f *FinancialHealth) EditableFilter() error {
	var errFields []string

	if f.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "cash_balance")
	}
	if f.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "burn_rate")
	}
	if f.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "cash_runway")
	}
	if f.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "burn_rate_change")
	}
	if f.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "quarterly_revenue")
	}
	if f.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "revenue_growth")
	}
	if f.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "gross_margin")
	}
	if f.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "net_margin")
	}
	if f.IsEditable&(1<<8) == 0 {
		errFields = append(errFields, "profitability_timeline")
	}
	if f.IsEditable&(1<<9) == 0 {
		errFields = append(errFields, "revenue_breakdowns")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
