package models

import (
	"fmt"

	"gorm.io/gorm"
)

type FundraisingStatus struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	LastRound             string `json:"last_round"`
	CurrentInvestors      string `json:"current_investors"`
	InvestorRelations     string `json:"investor_relations"`
	NextRound             string `json:"next_round"`
	TargetAmount          string `json:"target_amount"`
	InvestorPipeline      string `json:"investor_pipeline"`
	ValuationExpectations string `json:"valuation_expectations"`

	IsVisible  uint8 `gorm:"default:127" json:"-"`
	IsEditable uint8 `gorm:"default:127" json:"-"`
}

func (f *FundraisingStatus) TableName() string {
	return "fund"
}

func (f *FundraisingStatus) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"last_round",
		"current_investors",
		"investor_relations",
		"next_round",
		"target_amount",
		"investor_pipeline",
		"valuation_expectations",
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

// VisibilityFilter returns a map of visible fields based on IsVisible and fullAccess,
// using JSON field names as keys.
func (f *FundraisingStatus) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":                f.Version,
			"last_round":             f.LastRound,
			"current_investors":      f.CurrentInvestors,
			"investor_relations":     f.InvestorRelations,
			"next_round":             f.NextRound,
			"target_amount":          f.TargetAmount,
			"investor_pipeline":      f.InvestorPipeline,
			"valuation_expectations": f.ValuationExpectations,
		}
	}

	result := make(map[string]any)
	result["version"] = f.Version
	if f.IsVisible&(1<<0) != 0 {
		result["last_round"] = f.LastRound
	}
	if f.IsVisible&(1<<1) != 0 {
		result["current_investors"] = f.CurrentInvestors
	}
	if f.IsVisible&(1<<2) != 0 {
		result["investor_relations"] = f.InvestorRelations
	}
	if f.IsVisible&(1<<3) != 0 {
		result["next_round"] = f.NextRound
	}
	if f.IsVisible&(1<<4) != 0 {
		result["target_amount"] = f.TargetAmount
	}
	if f.IsVisible&(1<<5) != 0 {
		result["investor_pipeline"] = f.InvestorPipeline
	}
	if f.IsVisible&(1<<6) != 0 {
		result["valuation_expectations"] = f.ValuationExpectations
	}
	return result
}

func (f *FundraisingStatus) EditableFilter() error {
	var errFields []string

	if f.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "last_round")
	}
	if f.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "current_investors")
	}
	if f.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "investor_relations")
	}
	if f.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "next_round")
	}
	if f.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "target_amount")
	}
	if f.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "investor_pipeline")
	}
	if f.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "valuation_expectations")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
