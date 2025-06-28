package models

import (
	"fmt"

	"gorm.io/gorm"
)

type TeamPerformance struct {
	gorm.Model
	CompanyID              uint   `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID              uint   `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version                uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	TeamSize               string
	NewHires               string
	Turnover               string
	VacantPositions        string
	LeadershipAlignment    string
	TeamStrengths          string
	SkillGaps              string
	DevelopmentInitiatives string

	IsVisible  uint8 `gorm:"default:255"`
	IsEditable uint8 `gorm:"default:255"`
}

func (t *TeamPerformance) TableName() string {
	return "teamperf"
}

// Bit positions: 0 = TeamSize, 1 = NewHires, 2 = Turnover, 3 = VacantPositions, 4 = LeadershipAlignment,
// 5 = TeamStrengths, 6 = SkillGaps, 7 = DevelopmentInitiatives
func (t *TeamPerformance) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"TeamSize":               t.TeamSize,
			"NewHires":               t.NewHires,
			"Turnover":               t.Turnover,
			"VacantPositions":        t.VacantPositions,
			"LeadershipAlignment":    t.LeadershipAlignment,
			"TeamStrengths":          t.TeamStrengths,
			"SkillGaps":              t.SkillGaps,
			"DevelopmentInitiatives": t.DevelopmentInitiatives,
		}
	}

	result := make(map[string]any)
	if t.IsVisible&(1<<0) != 0 {
		result["TeamSize"] = t.TeamSize
	}
	if t.IsVisible&(1<<1) != 0 {
		result["NewHires"] = t.NewHires
	}
	if t.IsVisible&(1<<2) != 0 {
		result["Turnover"] = t.Turnover
	}
	if t.IsVisible&(1<<3) != 0 {
		result["VacantPositions"] = t.VacantPositions
	}
	if t.IsVisible&(1<<4) != 0 {
		result["LeadershipAlignment"] = t.LeadershipAlignment
	}
	if t.IsVisible&(1<<5) != 0 {
		result["TeamStrengths"] = t.TeamStrengths
	}
	if t.IsVisible&(1<<6) != 0 {
		result["SkillGaps"] = t.SkillGaps
	}
	if t.IsVisible&(1<<7) != 0 {
		result["DevelopmentInitiatives"] = t.DevelopmentInitiatives
	}
	return result
}

func (t *TeamPerformance) EditableFilter() error {
	var errFields []string

	if t.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "TeamSize")
	}
	if t.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "NewHires")
	}
	if t.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "Turnover")
	}
	if t.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "VacantPositions")
	}
	if t.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "LeadershipAlignment")
	}
	if t.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "TeamStrengths")
	}
	if t.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "SkillGaps")
	}
	if t.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "DevelopmentInitiatives")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
