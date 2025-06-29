package models

import (
	"fmt"

	"gorm.io/gorm"
)

type MarketTraction struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	NewCustomers        string `json:"new_customers"`
	TotalCustomers      string `json:"total_customers"`
	CustomerGrowth      string `json:"customer_growth"`
	RetentionRate       string `json:"retention_rate"`
	ChurnRate           string `json:"churn_rate"`
	PipelineValue       string `json:"pipeline_value"`
	ConversionRate      string `json:"conversion_rate"`
	SalesCycle          string `json:"sales_cycle"`
	SalesProcessChanges string `json:"sales_process_changes"`
	MarketShare         string `json:"market_share"`
	MarketShareChange   string `json:"market_share_change"`
	MarketTrends        string `json:"market_trends"`

	IsVisible  uint16 `gorm:"default:4095" json:"-"`
	IsEditable uint16 `gorm:"default:4095" json:"-"`
}

func (m *MarketTraction) TableName() string {
	return "market"
}

func (m *MarketTraction) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"new_customers",
		"total_customers",
		"customer_growth",
		"retention_rate",
		"churn_rate",
		"pipeline_value",
		"conversion_rate",
		"sales_cycle",
		"sales_process_changes",
		"market_share",
		"market_share_change",
		"market_trends",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if m.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

// VisibilityFilter returns a map of visible fields based on IsVisible and fullAccess, using JSON field names as keys.
func (m *MarketTraction) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":               m.Version,
			"new_customers":         m.NewCustomers,
			"total_customers":       m.TotalCustomers,
			"customer_growth":       m.CustomerGrowth,
			"retention_rate":        m.RetentionRate,
			"churn_rate":            m.ChurnRate,
			"pipeline_value":        m.PipelineValue,
			"conversion_rate":       m.ConversionRate,
			"sales_cycle":           m.SalesCycle,
			"sales_process_changes": m.SalesProcessChanges,
			"market_share":          m.MarketShare,
			"market_share_change":   m.MarketShareChange,
			"market_trends":         m.MarketTrends,
		}
	}

	result := make(map[string]any)
	result["version"] = m.Version
	if m.IsVisible&(1<<0) != 0 {
		result["new_customers"] = m.NewCustomers
	}
	if m.IsVisible&(1<<1) != 0 {
		result["total_customers"] = m.TotalCustomers
	}
	if m.IsVisible&(1<<2) != 0 {
		result["customer_growth"] = m.CustomerGrowth
	}
	if m.IsVisible&(1<<3) != 0 {
		result["retention_rate"] = m.RetentionRate
	}
	if m.IsVisible&(1<<4) != 0 {
		result["churn_rate"] = m.ChurnRate
	}
	if m.IsVisible&(1<<5) != 0 {
		result["pipeline_value"] = m.PipelineValue
	}
	if m.IsVisible&(1<<6) != 0 {
		result["conversion_rate"] = m.ConversionRate
	}
	if m.IsVisible&(1<<7) != 0 {
		result["sales_cycle"] = m.SalesCycle
	}
	if m.IsVisible&(1<<8) != 0 {
		result["sales_process_changes"] = m.SalesProcessChanges
	}
	if m.IsVisible&(1<<9) != 0 {
		result["market_share"] = m.MarketShare
	}
	if m.IsVisible&(1<<10) != 0 {
		result["market_share_change"] = m.MarketShareChange
	}
	if m.IsVisible&(1<<11) != 0 {
		result["market_trends"] = m.MarketTrends
	}
	return result
}

func (m *MarketTraction) EditableFilter() error {
	var errFields []string

	if m.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "new_customers")
	}
	if m.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "total_customers")
	}
	if m.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "customer_growth")
	}
	if m.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "retention_rate")
	}
	if m.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "churn_rate")
	}
	if m.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "pipeline_value")
	}
	if m.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "conversion_rate")
	}
	if m.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "sales_cycle")
	}
	if m.IsEditable&(1<<8) == 0 {
		errFields = append(errFields, "sales_process_changes")
	}
	if m.IsEditable&(1<<9) == 0 {
		errFields = append(errFields, "market_share")
	}
	if m.IsEditable&(1<<10) == 0 {
		errFields = append(errFields, "market_share_change")
	}
	if m.IsEditable&(1<<11) == 0 {
		errFields = append(errFields, "market_trends")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
