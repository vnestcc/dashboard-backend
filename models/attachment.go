package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Attachment struct {
	gorm.Model
	CompanyID            uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID            uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version              int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	FinancialStatements  string
	PitchDeck            string
	ProductRoadmap       string
	PerformanceDashboard string
	OrgChart             string

	IsVisible  int
	IsEditable int
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
