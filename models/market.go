package models

import (
	"fmt"

	"gorm.io/gorm"
)

type MarketTraction struct {
	gorm.Model
	CompanyID           uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID           uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version             int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	NewCustomers        string
	TotalCustomers      string
	CustomerGrowth      string
	RetentionRate       string
	ChurnRate           string
	PipelineValue       string
	ConversionRate      string
	SalesCycle          string
	SalesProcessChanges string
	MarketShare         string
	MarketShareChange   string
	MarketTrends        string

	IsVisible  int
	IsEditable int
}

func (m *MarketTraction) TableName() string {
	return "market"
}

// VisibilityFilter returns a map of visible fields based on IsVisible and fullAccess.
// Bit positions: 0 = NewCustomers, 1 = TotalCustomers, 2 = CustomerGrowth, 3 = RetentionRate, 4 = ChurnRate,
// 5 = PipelineValue, 6 = ConversionRate, 7 = SalesCycle, 8 = SalesProcessChanges, 9 = MarketShare,
// 10 = MarketShareChange, 11 = MarketTrends
func (m *MarketTraction) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"NewCustomers":        m.NewCustomers,
			"TotalCustomers":      m.TotalCustomers,
			"CustomerGrowth":      m.CustomerGrowth,
			"RetentionRate":       m.RetentionRate,
			"ChurnRate":           m.ChurnRate,
			"PipelineValue":       m.PipelineValue,
			"ConversionRate":      m.ConversionRate,
			"SalesCycle":          m.SalesCycle,
			"SalesProcessChanges": m.SalesProcessChanges,
			"MarketShare":         m.MarketShare,
			"MarketShareChange":   m.MarketShareChange,
			"MarketTrends":        m.MarketTrends,
		}
	}

	result := make(map[string]any)
	if m.IsVisible&(1<<0) != 0 {
		result["NewCustomers"] = m.NewCustomers
	}
	if m.IsVisible&(1<<1) != 0 {
		result["TotalCustomers"] = m.TotalCustomers
	}
	if m.IsVisible&(1<<2) != 0 {
		result["CustomerGrowth"] = m.CustomerGrowth
	}
	if m.IsVisible&(1<<3) != 0 {
		result["RetentionRate"] = m.RetentionRate
	}
	if m.IsVisible&(1<<4) != 0 {
		result["ChurnRate"] = m.ChurnRate
	}
	if m.IsVisible&(1<<5) != 0 {
		result["PipelineValue"] = m.PipelineValue
	}
	if m.IsVisible&(1<<6) != 0 {
		result["ConversionRate"] = m.ConversionRate
	}
	if m.IsVisible&(1<<7) != 0 {
		result["SalesCycle"] = m.SalesCycle
	}
	if m.IsVisible&(1<<8) != 0 {
		result["SalesProcessChanges"] = m.SalesProcessChanges
	}
	if m.IsVisible&(1<<9) != 0 {
		result["MarketShare"] = m.MarketShare
	}
	if m.IsVisible&(1<<10) != 0 {
		result["MarketShareChange"] = m.MarketShareChange
	}
	if m.IsVisible&(1<<11) != 0 {
		result["MarketTrends"] = m.MarketTrends
	}
	return result
}

func (m *MarketTraction) EditableFilter() error {
	var errFields []string

	if m.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "NewCustomers")
	}
	if m.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "TotalCustomers")
	}
	if m.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "CustomerGrowth")
	}
	if m.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "RetentionRate")
	}
	if m.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "ChurnRate")
	}
	if m.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "PipelineValue")
	}
	if m.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "ConversionRate")
	}
	if m.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "SalesCycle")
	}
	if m.IsEditable&(1<<8) == 0 {
		errFields = append(errFields, "SalesProcessChanges")
	}
	if m.IsEditable&(1<<9) == 0 {
		errFields = append(errFields, "MarketShare")
	}
	if m.IsEditable&(1<<10) == 0 {
		errFields = append(errFields, "MarketShareChange")
	}
	if m.IsEditable&(1<<11) == 0 {
		errFields = append(errFields, "MarketTrends")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
