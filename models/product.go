package models

import (
	"fmt"

	"gorm.io/gorm"
)

type ProductDevelopment struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	MilestonesAchieved  string `json:"milestones_achieved"`
	MilestonesMissed    string `json:"milestones_missed"`
	Roadmap             string `json:"roadmap"`
	ActiveUsers         string `json:"active_users"`
	EngagementMetrics   string `json:"engagement_metrics"`
	NPS                 string `json:"nps"`
	FeatureAdoption     string `json:"feature_adoption"`
	TechnicalChallenges string `json:"technical_challenges"`
	TechnicalDebt       string `json:"technical_debt"`
	ProductBottlenecks  string `json:"product_bottlenecks"`

	IsVisible  uint16 `gorm:"default:1023" json:"-"`
	IsEditable uint16 `gorm:"default:1023" json:"-"`
}

func (p *ProductDevelopment) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"MilestonesAchieved",
		"MilestonesMissed",
		"Roadmap",
		"ActiveUsers",
		"EngagementMetrics",
		"NPS",
		"FeatureAdoption",
		"TechnicalChallenges",
		"TechnicalDebt",
		"ProductBottlenecks",
	}
	if fullAccess {
		return fields
	}
	var visibleFields []string
	for i, field := range fields {
		if p.IsVisible&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

// Bit positions: 0 = MilestonesAchieved, 1 = MilestonesMissed, 2 = Roadmap, 3 = ActiveUsers, 4 = EngagementMetrics,
// 5 = NPS, 6 = FeatureAdoption, 7 = TechnicalChallenges, 8 = TechnicalDebt, 9 = ProductBottlenecks
func (p *ProductDevelopment) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"MilestonesAchieved":  p.MilestonesAchieved,
			"MilestonesMissed":    p.MilestonesMissed,
			"Roadmap":             p.Roadmap,
			"ActiveUsers":         p.ActiveUsers,
			"EngagementMetrics":   p.EngagementMetrics,
			"NPS":                 p.NPS,
			"FeatureAdoption":     p.FeatureAdoption,
			"TechnicalChallenges": p.TechnicalChallenges,
			"TechnicalDebt":       p.TechnicalDebt,
			"ProductBottlenecks":  p.ProductBottlenecks,
		}
	}

	result := make(map[string]any)
	if p.IsVisible&(1<<0) != 0 {
		result["MilestonesAchieved"] = p.MilestonesAchieved
	}
	if p.IsVisible&(1<<1) != 0 {
		result["MilestonesMissed"] = p.MilestonesMissed
	}
	if p.IsVisible&(1<<2) != 0 {
		result["Roadmap"] = p.Roadmap
	}
	if p.IsVisible&(1<<3) != 0 {
		result["ActiveUsers"] = p.ActiveUsers
	}
	if p.IsVisible&(1<<4) != 0 {
		result["EngagementMetrics"] = p.EngagementMetrics
	}
	if p.IsVisible&(1<<5) != 0 {
		result["NPS"] = p.NPS
	}
	if p.IsVisible&(1<<6) != 0 {
		result["FeatureAdoption"] = p.FeatureAdoption
	}
	if p.IsVisible&(1<<7) != 0 {
		result["TechnicalChallenges"] = p.TechnicalChallenges
	}
	if p.IsVisible&(1<<8) != 0 {
		result["TechnicalDebt"] = p.TechnicalDebt
	}
	if p.IsVisible&(1<<9) != 0 {
		result["ProductBottlenecks"] = p.ProductBottlenecks
	}
	return result
}

func (p *ProductDevelopment) EditableFilter() error {
	var errFields []string

	if p.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "MilestonesAchieved")
	}
	if p.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "MilestonesMissed")
	}
	if p.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "Roadmap")
	}
	if p.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "ActiveUsers")
	}
	if p.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "EngagementMetrics")
	}
	if p.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "NPS")
	}
	if p.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "FeatureAdoption")
	}
	if p.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "TechnicalChallenges")
	}
	if p.IsEditable&(1<<8) == 0 {
		errFields = append(errFields, "TechnicalDebt")
	}
	if p.IsEditable&(1<<9) == 0 {
		errFields = append(errFields, "ProductBottlenecks")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
