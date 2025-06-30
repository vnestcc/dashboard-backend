package models

import (
	"fmt"

	"gorm.io/gorm"
)

type AdditionalInfo struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	GrowthChallenges         string `json:"growth_challenges"`
	SupportNeeded            string `json:"support_needed"`
	PolicyChanges            string `json:"policy_changes"`
	PolicyImpact             string `json:"policy_impact"`
	MitigationStrategies     string `json:"mitigation_strategies"`
	NewInitiatives           string `json:"new_initiatives"`
	InitiativeProgress       string `json:"initiative_progress"`
	BusinessModelAdjustments string `json:"business_model_adjustments"`

	IsVisible  uint8 `gorm:"default:255"`
	IsEditable uint8 `gorm:"default:255"`
}

func (a *AdditionalInfo) EditableList() []string {
	fields := []string{
		"growth_challenges",
		"support_needed",
		"policy_changes",
		"policy_impact",
		"mitigation_strategies",
		"new_initiatives",
		"initiative_progress",
		"business_model_adjustments",
	}
	var visibleFields []string
	for i, field := range fields {
		if a.IsEditable&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (a *AdditionalInfo) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"growth_challenges",
		"support_needed",
		"policy_changes",
		"policy_impact",
		"mitigation_strategies",
		"new_initiatives",
		"initiative_progress",
		"business_model_adjustments",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if a.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (a *AdditionalInfo) TableName() string {
	return "additional"
}

func (a *AdditionalInfo) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":                    a.Version,
			"growth_challenges":          a.GrowthChallenges,
			"support_needed":             a.SupportNeeded,
			"policy_changes":             a.PolicyChanges,
			"policy_impact":              a.PolicyImpact,
			"mitigation_strategies":      a.MitigationStrategies,
			"new_initiatives":            a.NewInitiatives,
			"initiative_progress":        a.InitiativeProgress,
			"business_model_adjustments": a.BusinessModelAdjustments,
		}
	}

	result := make(map[string]any)
	result["version"] = a.Version
	if a.IsVisible&(1<<0) != 0 {
		result["growth_challenges"] = a.GrowthChallenges
	}
	if a.IsVisible&(1<<1) != 0 {
		result["support_needed"] = a.SupportNeeded
	}
	if a.IsVisible&(1<<2) != 0 {
		result["policy_changes"] = a.PolicyChanges
	}
	if a.IsVisible&(1<<3) != 0 {
		result["policy_impact"] = a.PolicyImpact
	}
	if a.IsVisible&(1<<4) != 0 {
		result["mitigation_strategies"] = a.MitigationStrategies
	}
	if a.IsVisible&(1<<5) != 0 {
		result["new_initiatives"] = a.NewInitiatives
	}
	if a.IsVisible&(1<<6) != 0 {
		result["initiative_progress"] = a.InitiativeProgress
	}
	if a.IsVisible&(1<<7) != 0 {
		result["business_model_adjustments"] = a.BusinessModelAdjustments
	}
	return result
}

func (a *AdditionalInfo) EditableFilter() error {
	var errFields []string

	if a.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "growth_challenges")
	}
	if a.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "support_needed")
	}
	if a.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "policy_changes")
	}
	if a.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "policy_impact")
	}
	if a.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "mitigation_strategies")
	}
	if a.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "new_initiatives")
	}
	if a.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "initiative_progress")
	}
	if a.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "business_model_adjustments")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
