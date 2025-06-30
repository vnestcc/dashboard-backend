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
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	FinancialRating   int        `json:"financial_rating"`   // 1-10 bit 0
	MarketRating      int        `json:"market_rating"`      // 1-10
	ProductRating     int        `json:"product_rating"`     // 1-10
	TeamRating        int        `json:"team_rating"`        // 1-10
	OperationalRating int        `json:"operational_rating"` // 1-10
	OverallRating     int        `json:"overall_rating"`     // 1-10
	Priorities        Priorities `gorm:"type:text" json:"priorities"`
	IncubatorSupport  string     `json:"incubator_support"` // bit 7

	IsVisible  uint8 `gorm:"default:255" json:"-"`
	IsEditable uint8 `gorm:"default:255" json:"-"` // usually hidden from response
}

func (s *SelfAssessment) TableName() string {
	return "assessment"
}

func (s *SelfAssessment) EditableList() []string {
	fields := []string{
		"financial_rating",
		"market_rating",
		"product_rating",
		"team_rating",
		"operational_rating",
		"overall_rating",
		"priorities",
		"incubator_support",
	}
	var visibleFields []string
	for i, field := range fields {
		if s.IsEditable&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (s *SelfAssessment) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"financial_rating",
		"market_rating",
		"product_rating",
		"team_rating",
		"operational_rating",
		"overall_rating",
		"priorities",
		"incubator_support",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if s.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (s *SelfAssessment) BeforeSave(tx *gorm.DB) (err error) {
	ratings := []struct {
		value int
		name  string
	}{
		{s.FinancialRating, "financial_rating"},
		{s.MarketRating, "market_rating"},
		{s.ProductRating, "product_rating"},
		{s.TeamRating, "team_rating"},
		{s.OperationalRating, "operational_rating"},
		{s.OverallRating, "overall_rating"},
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
			"version":            s.Version,
			"financial_rating":   s.FinancialRating,
			"market_rating":      s.MarketRating,
			"product_rating":     s.ProductRating,
			"team_rating":        s.TeamRating,
			"operational_rating": s.OperationalRating,
			"overall_rating":     s.OverallRating,
			"priorities":         s.Priorities,
			"incubator_support":  s.IncubatorSupport,
		}
	}

	result := make(map[string]any)
	result["version"] = s.Version
	if s.IsVisible&(1<<0) != 0 {
		result["financial_rating"] = s.FinancialRating
	}
	if s.IsVisible&(1<<1) != 0 {
		result["market_rating"] = s.MarketRating
	}
	if s.IsVisible&(1<<2) != 0 {
		result["product_rating"] = s.ProductRating
	}
	if s.IsVisible&(1<<3) != 0 {
		result["team_rating"] = s.TeamRating
	}
	if s.IsVisible&(1<<4) != 0 {
		result["operational_rating"] = s.OperationalRating
	}
	if s.IsVisible&(1<<5) != 0 {
		result["overall_rating"] = s.OverallRating
	}
	if s.IsVisible&(1<<6) != 0 {
		result["priorities"] = s.Priorities
	}
	if s.IsVisible&(1<<7) != 0 {
		result["incubator_support"] = s.IncubatorSupport
	}
	return result
}

func (s *SelfAssessment) EditableFilter() error {
	var errFields []string

	if s.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "financial_rating")
	}
	if s.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "market_rating")
	}
	if s.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "product_rating")
	}
	if s.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "team_rating")
	}
	if s.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "operational_rating")
	}
	if s.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "overall_rating")
	}
	if s.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "priorities")
	}
	if s.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "incubator_support")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
