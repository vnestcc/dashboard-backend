package models

import (
	"database/sql/driver"
	"errors"
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
	FinancialRating   int        // 1-10
	MarketRating      int        // 1-10
	ProductRating     int        // 1-10
	TeamRating        int        // 1-10
	OperationalRating int        // 1-10
	OverallRating     int        // 1-10
	Priorities        Priorities `gorm:"type:text"`
	IncubatorSupport  string

	IsVisible  int
	IsEditable int
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
