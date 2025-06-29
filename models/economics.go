package models

import (
	"fmt"

	"gorm.io/gorm"
)

type UnitEconomics struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	CAC        string `json:"cac"`
	CACChange  string `json:"cac_change"`
	LTV        string `json:"ltv"`
	LTVRatio   string `json:"ltv_ratio"`
	CACPayback string `json:"cac_payback"`
	ARPU       string `json:"arpu"`

	MarketingBreakdowns []MarketingBreakdown `json:"marketing_breakdowns"`

	IsVisible  uint8 `gorm:"default:127" json:"-"`
	IsEditable uint8 `gorm:"default:127" json:"-"`
}

type MarketingBreakdown struct {
	gorm.Model
	UnitEconomicsID uint   `json:"-"`
	Channel         string `json:"channel"`
	Spend           string `json:"spend"`
	Budget          string `json:"budget"`
	CAC             string `json:"cac"`
}

func (u *UnitEconomics) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"CAC",
		"CACChange",
		"LTV",
		"LTVRatio",
		"CACPayback",
		"ARPU",
		"MarketingBreakdowns",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if u.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (u *UnitEconomics) TableName() string {
	return "economics"
}

func (u *UnitEconomics) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"CAC":                 u.CAC,
			"CACChange":           u.CACChange,
			"LTV":                 u.LTV,
			"LTVRatio":            u.LTVRatio,
			"CACPayback":          u.CACPayback,
			"ARPU":                u.ARPU,
			"MarketingBreakdowns": u.MarketingBreakdowns,
		}
	}

	result := make(map[string]any)
	if u.IsVisible&(1<<0) != 0 {
		result["CAC"] = u.CAC
	}
	if u.IsVisible&(1<<1) != 0 {
		result["CACChange"] = u.CACChange
	}
	if u.IsVisible&(1<<2) != 0 {
		result["LTV"] = u.LTV
	}
	if u.IsVisible&(1<<3) != 0 {
		result["LTVRatio"] = u.LTVRatio
	}
	if u.IsVisible&(1<<4) != 0 {
		result["CACPayback"] = u.CACPayback
	}
	if u.IsVisible&(1<<5) != 0 {
		result["ARPU"] = u.ARPU
	}
	if u.IsVisible&(1<<6) != 0 {
		result["MarketingBreakdowns"] = u.MarketingBreakdowns
	}
	return result
}

// EditableFilter returns an error listing all fields that are not editable.
// Bit mapping:
// 0: CAC
// 1: CACChange
// 2: LTV
// 3: LTVRatio
// 4: CACPayback
// 5: ARPU
// 6: MarketingBreakdowns (optional, add if you want bit-level control for slice fields)
func (u *UnitEconomics) EditableFilter() error {
	var errFields []string

	if u.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "CAC")
	}
	if u.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "CACChange")
	}
	if u.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "LTV")
	}
	if u.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "LTVRatio")
	}
	if u.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "CACPayback")
	}
	if u.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "ARPU")
	}
	if u.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "MarketingBreakdowns")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
