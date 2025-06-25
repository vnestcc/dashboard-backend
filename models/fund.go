package models

import (
	"fmt"

	"gorm.io/gorm"
)

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

	IsVisible  int `gorm:"default:127"`
	IsEditable int `gorm:"default:127"`
}

func (f *FundraisingStatus) TableName() string {
	return "fund"
}

// VisibilityFilter returns a map of visible fields based on IsVisible and fullAccess.
// Bit positions: 0 = LastRound, 1 = CurrentInvestors, 2 = InvestorRelations, 3 = NextRound,
// 4 = TargetAmount, 5 = InvestorPipeline, 6 = ValuationExpectations
func (f *FundraisingStatus) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"LastRound":             f.LastRound,
			"CurrentInvestors":      f.CurrentInvestors,
			"InvestorRelations":     f.InvestorRelations,
			"NextRound":             f.NextRound,
			"TargetAmount":          f.TargetAmount,
			"InvestorPipeline":      f.InvestorPipeline,
			"ValuationExpectations": f.ValuationExpectations,
		}
	}

	result := make(map[string]any)
	if f.IsVisible&(1<<0) != 0 {
		result["LastRound"] = f.LastRound
	}
	if f.IsVisible&(1<<1) != 0 {
		result["CurrentInvestors"] = f.CurrentInvestors
	}
	if f.IsVisible&(1<<2) != 0 {
		result["InvestorRelations"] = f.InvestorRelations
	}
	if f.IsVisible&(1<<3) != 0 {
		result["NextRound"] = f.NextRound
	}
	if f.IsVisible&(1<<4) != 0 {
		result["TargetAmount"] = f.TargetAmount
	}
	if f.IsVisible&(1<<5) != 0 {
		result["InvestorPipeline"] = f.InvestorPipeline
	}
	if f.IsVisible&(1<<6) != 0 {
		result["ValuationExpectations"] = f.ValuationExpectations
	}
	return result
}

func (f *FundraisingStatus) EditableFilter() error {
	var errFields []string

	if f.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "LastRound")
	}
	if f.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "CurrentInvestors")
	}
	if f.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "InvestorRelations")
	}
	if f.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "NextRound")
	}
	if f.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "TargetAmount")
	}
	if f.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "InvestorPipeline")
	}
	if f.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "ValuationExpectations")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
