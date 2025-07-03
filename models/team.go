package models

import (
	"fmt"

	"gorm.io/gorm"
)

type TeamPerformance struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	TeamSize               string   `json:"team_size"`
	NewHires               string   `json:"new_hires"`
	Turnover               string   `json:"turnover"`
	VacantPositions        string   `json:"vacant_positions"`
	LeadershipAlignment    string   `json:"leadership_alignment"`
	TeamStrengths          []string `json:"team_strengths"`
	SkillGaps              string   `json:"skill_gaps"`
	DevelopmentInitiatives []string `json:"development_initiatives"`

	IsVisible  uint8 `gorm:"default:255" json:"-"`
	IsEditable uint8 `gorm:"default:255" json:"-"`
}

func (t *TeamPerformance) TableName() string {
	return "teamperf"
}

func (t *TeamPerformance) EditableList() []string {
	fields := []string{
		"team_size",
		"new_hires",
		"turnover",
		"vacant_positions",
		"leadership_alignment",
		"team_strengths",
		"skill_gaps",
		"development_initiatives",
	}
	var visibleFields []string
	for i, field := range fields {
		if t.IsEditable&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (t *TeamPerformance) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"team_size",
		"new_hires",
		"turnover",
		"vacant_positions",
		"leadership_alignment",
		"team_strengths",
		"skill_gaps",
		"development_initiatives",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if t.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

// Bit positions: 0 = team_size, 1 = new_hires, 2 = turnover, 3 = vacant_positions, 4 = leadership_alignment,
// 5 = team_strengths, 6 = skill_gaps, 7 = development_initiatives
func (t *TeamPerformance) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":                 t.Version,
			"team_size":               t.TeamSize,
			"new_hires":               t.NewHires,
			"turnover":                t.Turnover,
			"vacant_positions":        t.VacantPositions,
			"leadership_alignment":    t.LeadershipAlignment,
			"team_strengths":          t.TeamStrengths,
			"skill_gaps":              t.SkillGaps,
			"development_initiatives": t.DevelopmentInitiatives,
		}
	}

	result := make(map[string]any)
	result["version"] = t.Version
	if t.IsVisible&(1<<0) != 0 {
		result["team_size"] = t.TeamSize
	}
	if t.IsVisible&(1<<1) != 0 {
		result["new_hires"] = t.NewHires
	}
	if t.IsVisible&(1<<2) != 0 {
		result["turnover"] = t.Turnover
	}
	if t.IsVisible&(1<<3) != 0 {
		result["vacant_positions"] = t.VacantPositions
	}
	if t.IsVisible&(1<<4) != 0 {
		result["leadership_alignment"] = t.LeadershipAlignment
	}
	if t.IsVisible&(1<<5) != 0 {
		result["team_strengths"] = t.TeamStrengths
	}
	if t.IsVisible&(1<<6) != 0 {
		result["skill_gaps"] = t.SkillGaps
	}
	if t.IsVisible&(1<<7) != 0 {
		result["development_initiatives"] = t.DevelopmentInitiatives
	}
	return result
}

func (t *TeamPerformance) EditableFilter() error {
	var errFields []string

	if t.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "team_size")
	}
	if t.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "new_hires")
	}
	if t.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "turnover")
	}
	if t.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "vacant_positions")
	}
	if t.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "leadership_alignment")
	}
	if t.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "team_strengths")
	}
	if t.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "skill_gaps")
	}
	if t.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "development_initiatives")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
