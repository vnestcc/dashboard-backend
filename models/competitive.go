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
		"new_competitors",
		"competitor_strategies",
		"market_shifts",
		"differentiators",
		"threats",
		"defensive_strategies",
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
			"version":               c.Version,
			"new_competitors":       c.NewCompetitors,
			"competitor_strategies": c.CompetitorStrategies,
			"market_shifts":         c.MarketShifts,
			"differentiators":       c.Differentiators,
			"threats":               c.Threats,
			"defensive_strategies":  c.DefensiveStrategies,
		}
	}

	result := make(map[string]any)
	result["version"] = c.Version
	if c.IsVisible&(1<<0) != 0 {
		result["new_competitors"] = c.NewCompetitors
	}
	if c.IsVisible&(1<<1) != 0 {
		result["competitor_strategies"] = c.CompetitorStrategies
	}
	if c.IsVisible&(1<<2) != 0 {
		result["market_shifts"] = c.MarketShifts
	}
	if c.IsVisible&(1<<3) != 0 {
		result["differentiators"] = c.Differentiators
	}
	if c.IsVisible&(1<<4) != 0 {
		result["threats"] = c.Threats
	}
	if c.IsVisible&(1<<5) != 0 {
		result["defensive_strategies"] = c.DefensiveStrategies
	}
	return result
}

func (c *CompetitiveLandscape) EditableFilter() error {
	var errFields []string

	if c.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "new_competitors")
	}
	if c.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "competitor_strategies")
	}
	if c.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "market_shifts")
	}
	if c.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "differentiators")
	}
	if c.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "threats")
	}
	if c.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "defensive_strategies")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
