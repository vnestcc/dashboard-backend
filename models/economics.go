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
	gorm.Model      `json:"-"`
	UnitEconomicsID uint   `json:"-"`
	Channel         string `json:"channel"`
	Spend           string `json:"spend"`
	Budget          string `json:"budget"`
	CAC             string `json:"cac"`
}

func (u *UnitEconomics) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"cac",
		"cac_change",
		"ltv",
		"ltv_ratio",
		"cac_payback",
		"arpu",
		"marketing_breakdowns",
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
			"version":              u.Version,
			"cac":                  u.CAC,
			"cac_change":           u.CACChange,
			"ltv":                  u.LTV,
			"ltv_ratio":            u.LTVRatio,
			"cac_payback":          u.CACPayback,
			"arpu":                 u.ARPU,
			"marketing_breakdowns": u.MarketingBreakdowns,
		}
	}

	result := make(map[string]any)
	result["version"] = u.Version
	if u.IsVisible&(1<<0) != 0 {
		result["cac"] = u.CAC
	}
	if u.IsVisible&(1<<1) != 0 {
		result["cac_change"] = u.CACChange
	}
	if u.IsVisible&(1<<2) != 0 {
		result["ltv"] = u.LTV
	}
	if u.IsVisible&(1<<3) != 0 {
		result["ltv_ratio"] = u.LTVRatio
	}
	if u.IsVisible&(1<<4) != 0 {
		result["cac_payback"] = u.CACPayback
	}
	if u.IsVisible&(1<<5) != 0 {
		result["arpu"] = u.ARPU
	}
	if u.IsVisible&(1<<6) != 0 {
		result["marketing_breakdowns"] = u.MarketingBreakdowns
	}
	return result
}

// EditableFilter returns an error listing all fields that are not editable.
// Bit mapping:
// 0: cac
// 1: cac_change
// 2: ltv
// 3: ltv_ratio
// 4: cac_payback
// 5: arpu
// 6: marketing_breakdowns
func (u *UnitEconomics) EditableFilter() error {
	var errFields []string

	if u.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "cac")
	}
	if u.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "cac_change")
	}
	if u.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "ltv")
	}
	if u.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "ltv_ratio")
	}
	if u.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "cac_payback")
	}
	if u.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "arpu")
	}
	if u.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "marketing_breakdowns")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
