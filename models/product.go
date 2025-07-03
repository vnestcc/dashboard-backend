package models

import (
	"fmt"

	"gorm.io/gorm"
)

type ProductDevelopment struct {
	gorm.Model
	CompanyID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	QuarterID uint   `gorm:"not null;index:idx_unique_comp_quarter_version"`
	Version   uint32 `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`

	MilestonesAchieved  string   `json:"milestones_achieved"`
	MilestonesMissed    string   `json:"milestones_missed"`
	Roadmap             []string `json:"roadmap"` // array of size 3
	ActiveUsers         string   `json:"active_users"`
	EngagementMetrics   string   `json:"engagement_metrics"`
	NPS                 string   `json:"nps"`
	FeatureAdoption     string   `json:"feature_adoption"`
	TechnicalChallenges []string `json:"technical_challenges"` // array of size 3
	TechnicalDebt       string   `json:"technical_debt"`
	ProductBottlenecks  []string `json:"product_bottlenecks"` // array of size 3

	IsVisible  uint16 `gorm:"default:1023" json:"-"`
	IsEditable uint16 `gorm:"default:1023" json:"-"`
}

func (p *ProductDevelopment) TableName() string {
	return "product"
}

func (p *ProductDevelopment) EditableList() []string {
	fields := []string{
		"milestones_achieved",
		"milestones_missed",
		"roadmap",
		"active_users",
		"engagement_metrics",
		"nps",
		"feature_adoption",
		"technical_challenges",
		"technical_debt",
		"product_bottlenecks",
	}
	var visibleFields []string
	for i, field := range fields {
		if p.IsEditable&(1<<i) != 0 {
			visibleFields = append(visibleFields, field)
		}
	}
	return visibleFields
}

func (p *ProductDevelopment) VisibilityList(fullAccess bool) []string {
	fields := []string{
		"milestones_achieved",
		"milestones_missed",
		"roadmap",
		"active_users",
		"engagement_metrics",
		"nps",
		"feature_adoption",
		"technical_challenges",
		"technical_debt",
		"product_bottlenecks",
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

// Bit positions: 0 = milestones_achieved, 1 = milestones_missed, 2 = roadmap, 3 = active_users, 4 = engagement_metrics,
// 5 = nps, 6 = feature_adoption, 7 = technical_challenges, 8 = technical_debt, 9 = product_bottlenecks
func (p *ProductDevelopment) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"version":              p.Version,
			"milestones_achieved":  p.MilestonesAchieved,
			"milestones_missed":    p.MilestonesMissed,
			"roadmap":              p.Roadmap,
			"active_users":         p.ActiveUsers,
			"engagement_metrics":   p.EngagementMetrics,
			"nps":                  p.NPS,
			"feature_adoption":     p.FeatureAdoption,
			"technical_challenges": p.TechnicalChallenges,
			"technical_debt":       p.TechnicalDebt,
			"product_bottlenecks":  p.ProductBottlenecks,
		}
	}

	result := make(map[string]any)
	result["version"] = p.Version
	if p.IsVisible&(1<<0) != 0 {
		result["milestones_achieved"] = p.MilestonesAchieved
	}
	if p.IsVisible&(1<<1) != 0 {
		result["milestones_missed"] = p.MilestonesMissed
	}
	if p.IsVisible&(1<<2) != 0 {
		result["roadmap"] = p.Roadmap
	}
	if p.IsVisible&(1<<3) != 0 {
		result["active_users"] = p.ActiveUsers
	}
	if p.IsVisible&(1<<4) != 0 {
		result["engagement_metrics"] = p.EngagementMetrics
	}
	if p.IsVisible&(1<<5) != 0 {
		result["nps"] = p.NPS
	}
	if p.IsVisible&(1<<6) != 0 {
		result["feature_adoption"] = p.FeatureAdoption
	}
	if p.IsVisible&(1<<7) != 0 {
		result["technical_challenges"] = p.TechnicalChallenges
	}
	if p.IsVisible&(1<<8) != 0 {
		result["technical_debt"] = p.TechnicalDebt
	}
	if p.IsVisible&(1<<9) != 0 {
		result["product_bottlenecks"] = p.ProductBottlenecks
	}
	return result
}

func (p *ProductDevelopment) EditableFilter() error {
	var errFields []string

	if p.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "milestones_achieved")
	}
	if p.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "milestones_missed")
	}
	if p.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "roadmap")
	}
	if p.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "active_users")
	}
	if p.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "engagement_metrics")
	}
	if p.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "nps")
	}
	if p.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "feature_adoption")
	}
	if p.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "technical_challenges")
	}
	if p.IsEditable&(1<<8) == 0 {
		errFields = append(errFields, "technical_debt")
	}
	if p.IsEditable&(1<<9) == 0 {
		errFields = append(errFields, "product_bottlenecks")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
