package models

import (
	"fmt"

	"gorm.io/gorm"
)

type AdditionalInfo struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
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

func (a *AdditionalInfo) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"GrowthChallenges",
		"SupportNeeded",
		"PolicyChanges",
		"PolicyImpact",
		"MitigationStrategies",
		"NewInitiatives",
		"InitiativeProgress",
		"BusinessModelAdjustments",
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
			"GrowthChallenges":         a.GrowthChallenges,
			"SupportNeeded":            a.SupportNeeded,
			"PolicyChanges":            a.PolicyChanges,
			"PolicyImpact":             a.PolicyImpact,
			"MitigationStrategies":     a.MitigationStrategies,
			"NewInitiatives":           a.NewInitiatives,
			"InitiativeProgress":       a.InitiativeProgress,
			"BusinessModelAdjustments": a.BusinessModelAdjustments,
		}
	}

	result := make(map[string]any)
	if a.IsVisible&(1<<0) != 0 {
		result["GrowthChallenges"] = a.GrowthChallenges
	}
	if a.IsVisible&(1<<1) != 0 {
		result["SupportNeeded"] = a.SupportNeeded
	}
	if a.IsVisible&(1<<2) != 0 {
		result["PolicyChanges"] = a.PolicyChanges
	}
	if a.IsVisible&(1<<3) != 0 {
		result["PolicyImpact"] = a.PolicyImpact
	}
	if a.IsVisible&(1<<4) != 0 {
		result["MitigationStrategies"] = a.MitigationStrategies
	}
	if a.IsVisible&(1<<5) != 0 {
		result["NewInitiatives"] = a.NewInitiatives
	}
	if a.IsVisible&(1<<6) != 0 {
		result["InitiativeProgress"] = a.InitiativeProgress
	}
	if a.IsVisible&(1<<7) != 0 {
		result["BusinessModelAdjustments"] = a.BusinessModelAdjustments
	}
	return result
}

func (a *AdditionalInfo) EditableFilter() error {
	var errFields []string

	if a.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "GrowthChallenges")
	}
	if a.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "SupportNeeded")
	}
	if a.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "PolicyChanges")
	}
	if a.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "PolicyImpact")
	}
	if a.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "MitigationStrategies")
	}
	if a.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "NewInitiatives")
	}
	if a.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "InitiativeProgress")
	}
	if a.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "BusinessModelAdjustments")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)

	}
	return nil
}
