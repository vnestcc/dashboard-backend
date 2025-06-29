package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Attachment struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	FinancialStatements  string `json:"financial_statements"` // bit 0
	PitchDeck            string `json:"pitch_deck"`           // bit 1
	ProductRoadmap       string `json:"product_roadmap"`
	PerformanceDashboard string `json:"performance_dashboard"`
	OrgChart             string `json:"org_chart"` // bit 4

	IsVisible  uint8 `gorm:"default:31" json:"-"`
	IsEditable uint8 `gorm:"default:31" json:"-"`
}

func (a *Attachment) TableName() string {
	return "attachment"
}

func (a *Attachment) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"financial_statements",
		"pitch_deck",
		"product_roadmap",
		"performance_dashboard",
		"org_chart",
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

func (a *Attachment) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":               a.Version,
			"financial_statements":  a.FinancialStatements,
			"pitch_deck":            a.PitchDeck,
			"product_roadmap":       a.ProductRoadmap,
			"performance_dashboard": a.PerformanceDashboard,
			"org_chart":             a.OrgChart,
		}
	}

	result := make(map[string]any)
	result["version"] = a.Version
	if a.IsVisible&(1<<0) != 0 {
		result["financial_statements"] = a.FinancialStatements
	}
	if a.IsVisible&(1<<1) != 0 {
		result["pitch_deck"] = a.PitchDeck
	}
	if a.IsVisible&(1<<2) != 0 {
		result["product_roadmap"] = a.ProductRoadmap
	}
	if a.IsVisible&(1<<3) != 0 {
		result["performance_dashboard"] = a.PerformanceDashboard
	}
	if a.IsVisible&(1<<4) != 0 {
		result["org_chart"] = a.OrgChart
	}
	return result
}

func (a *Attachment) EditableFilter() error {
	var errFields []string

	if a.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "financial_statements")
	}
	if a.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "pitch_deck")
	}
	if a.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "product_roadmap")
	}
	if a.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "performance_dashboard")
	}
	if a.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "org_chart")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
