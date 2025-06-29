package models

import (
	"fmt"

	"gorm.io/gorm"
)

type CompetitiveLandscape struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	NewCompetitors       string `json:"new_competitors"`
	CompetitorStrategies string `json:"competitor_strategies"`
	MarketShifts         string `json:"market_shifts"`
	Differentiators      string `json:"differentiators"`
	Threats              string `json:"threats"`
	DefensiveStrategies  string `json:"defensive_strategies"`

	IsVisible  uint8 `gorm:"default:63" json:"-"`
	IsEditable uint8 `gorm:"default:63" json:"-"`
}

func (c *CompetitiveLandscape) TableName() string {
	return "competitive"
}

func (c *CompetitiveLandscape) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"NewCompetitors",
		"CompetitorStrategies",
		"MarketShifts",
		"Differentiators",
		"Threats",
		"DefensiveStrategies",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if c.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (c *CompetitiveLandscape) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"NewCompetitors":       c.NewCompetitors,
			"CompetitorStrategies": c.CompetitorStrategies,
			"MarketShifts":         c.MarketShifts,
			"Differentiators":      c.Differentiators,
			"Threats":              c.Threats,
			"DefensiveStrategies":  c.DefensiveStrategies,
		}
	}

	result := make(map[string]any)
	if c.IsVisible&(1<<0) != 0 {
		result["NewCompetitors"] = c.NewCompetitors
	}
	if c.IsVisible&(1<<1) != 0 {
		result["CompetitorStrategies"] = c.CompetitorStrategies
	}
	if c.IsVisible&(1<<2) != 0 {
		result["MarketShifts"] = c.MarketShifts
	}
	if c.IsVisible&(1<<3) != 0 {
		result["Differentiators"] = c.Differentiators
	}
	if c.IsVisible&(1<<4) != 0 {
		result["Threats"] = c.Threats
	}
	if c.IsVisible&(1<<5) != 0 {
		result["DefensiveStrategies"] = c.DefensiveStrategies
	}
	return result
}

func (c *CompetitiveLandscape) EditableFilter() error {
	var errFields []string

	if c.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "NewCompetitors")
	}
	if c.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "CompetitorStrategies")
	}
	if c.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "MarketShifts")
	}
	if c.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "Differentiators")
	}
	if c.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "Threats")
	}
	if c.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "DefensiveStrategies")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
