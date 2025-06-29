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
		"FinancialStatements",
		"PitchDeck",
		"ProductRoadmap",
		"PerformanceDashboard",
		"OrgChart",
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
			"FinancialStatements":  a.FinancialStatements,
			"PitchDeck":            a.PitchDeck,
			"ProductRoadmap":       a.ProductRoadmap,
			"PerformanceDashboard": a.PerformanceDashboard,
			"OrgChart":             a.OrgChart,
		}
	}

	result := make(map[string]any)
	if a.IsVisible&(1<<0) != 0 {
		result["FinancialStatements"] = a.FinancialStatements
	}
	if a.IsVisible&(1<<1) != 0 {
		result["PitchDeck"] = a.PitchDeck
	}
	if a.IsVisible&(1<<2) != 0 {
		result["ProductRoadmap"] = a.ProductRoadmap
	}
	if a.IsVisible&(1<<3) != 0 {
		result["PerformanceDashboard"] = a.PerformanceDashboard
	}
	if a.IsVisible&(1<<4) != 0 {
		result["OrgChart"] = a.OrgChart
	}
	return result
}

func (a *Attachment) EditableFilter() error {
	var errFields []string

	if a.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "FinancialStatements")
	}
	if a.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "PitchDeck")
	}
	if a.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "ProductRoadmap")
	}
	if a.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "PerformanceDashboard")
	}
	if a.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "OrgChart")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
