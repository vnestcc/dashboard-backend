package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Priorities []string

func (p Priorities) Value() (driver.Value, error) {
	return strings.Join(p, "|"), nil
}

func (p *Priorities) Scan(value any) error {
	if value == nil {
		*p = Priorities{}
		return nil
	}
	str, ok := value.(string)
	if !ok {
		bs, ok := value.([]byte)
		if !ok {
			return errors.New("failed to scan Priorities: type assertion to string or []byte failed")
		}
		str = string(bs)
	}
	if str == "" {
		*p = Priorities{}
		return nil
	}
	*p = strings.Split(str, "|")
	return nil
}

type SelfAssessment struct {
	gorm.Model
	CompanyID         uint       `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID         uint       `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version           int        `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	FinancialRating   int        // 1-10 bit 0
	MarketRating      int        // 1-10
	ProductRating     int        // 1-10
	TeamRating        int        // 1-10
	OperationalRating int        // 1-10
	OverallRating     int        // 1-10
	Priorities        Priorities `gorm:"type:text"`
	IncubatorSupport  string     // bit 7

	IsVisible  int `gorm:"default:255"`
	IsEditable int `gorm:"default:255"`
}

func (s *SelfAssessment) BeforeSave(tx *gorm.DB) (err error) {
	ratings := []struct {
		value int
		name  string
	}{
		{s.FinancialRating, "FinancialRating"},
		{s.MarketRating, "MarketRating"},
		{s.ProductRating, "ProductRating"},
		{s.TeamRating, "TeamRating"},
		{s.OperationalRating, "OperationalRating"},
		{s.OverallRating, "OverallRating"},
	}
	for _, r := range ratings {
		if r.value < 1 || r.value > 10 {
			return errors.New(r.name + " must be between 1 and 10")
		}
	}
	return nil
}

func (s *SelfAssessment) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"FinancialRating":   s.FinancialRating,
			"MarketRating":      s.MarketRating,
			"ProductRating":     s.ProductRating,
			"TeamRating":        s.TeamRating,
			"OperationalRating": s.OperationalRating,
			"OverallRating":     s.OverallRating,
			"Priorities":        s.Priorities,
			"IncubatorSupport":  s.IncubatorSupport,
		}
	}

	result := make(map[string]any)
	if s.IsVisible&(1<<0) != 0 {
		result["FinancialRating"] = s.FinancialRating
	}
	if s.IsVisible&(1<<1) != 0 {
		result["MarketRating"] = s.MarketRating
	}
	if s.IsVisible&(1<<2) != 0 {
		result["ProductRating"] = s.ProductRating
	}
	if s.IsVisible&(1<<3) != 0 {
		result["TeamRating"] = s.TeamRating
	}
	if s.IsVisible&(1<<4) != 0 {
		result["OperationalRating"] = s.OperationalRating
	}
	if s.IsVisible&(1<<5) != 0 {
		result["OverallRating"] = s.OverallRating
	}
	if s.IsVisible&(1<<6) != 0 {
		result["Priorities"] = s.Priorities
	}
	if s.IsVisible&(1<<7) != 0 {
		result["IncubatorSupport"] = s.IncubatorSupport
	}
	return result
}

func (s *SelfAssessment) EditableFilter() error {
	var errFields []string

	if s.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "FinancialRating")
	}
	if s.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "MarketRating")
	}
	if s.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "ProductRating")
	}
	if s.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "TeamRating")
	}
	if s.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "OperationalRating")
	}
	if s.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "OverallRating")
	}
	if s.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "Priorities")
	}
	if s.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "IncubatorSupport")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
