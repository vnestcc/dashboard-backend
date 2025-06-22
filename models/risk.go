package models

import (
	"fmt"

	"gorm.io/gorm"
)

type RiskManagement struct {
	gorm.Model
	CompanyID          uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	QuarterID          uint `gorm:"not null;index:idx_unique_comp_quarter_version,unique"`
	Version            int  `gorm:"not null;index:idx_unique_comp_quarter_version,unique;default:1"`
	RegulatoryChanges  string
	ComplianceStatus   string
	RegulatoryConcerns string
	SecurityAudits     string
	DataProtection     string
	SecurityIncidents  string
	KeyDependencies    string
	ContingencyPlans   string

	IsVisible  int `gorm:"default:255"`
	IsEditable int `gorm:"default:255"`
}

// Bit positions: 0 = RegulatoryChanges, 1 = ComplianceStatus, 2 = RegulatoryConcerns, 3 = SecurityAudits,
// 4 = DataProtection, 5 = SecurityIncidents, 6 = KeyDependencies, 7 = ContingencyPlans
func (r *RiskManagement) VisibilityFilter(fullAccess bool) map[string]any {
	if fullAccess {
		return map[string]any{
			"RegulatoryChanges":  r.RegulatoryChanges,
			"ComplianceStatus":   r.ComplianceStatus,
			"RegulatoryConcerns": r.RegulatoryConcerns,
			"SecurityAudits":     r.SecurityAudits,
			"DataProtection":     r.DataProtection,
			"SecurityIncidents":  r.SecurityIncidents,
			"KeyDependencies":    r.KeyDependencies,
			"ContingencyPlans":   r.ContingencyPlans,
		}
	}

	result := make(map[string]any)
	if r.IsVisible&(1<<0) != 0 {
		result["RegulatoryChanges"] = r.RegulatoryChanges
	}
	if r.IsVisible&(1<<1) != 0 {
		result["ComplianceStatus"] = r.ComplianceStatus
	}
	if r.IsVisible&(1<<2) != 0 {
		result["RegulatoryConcerns"] = r.RegulatoryConcerns
	}
	if r.IsVisible&(1<<3) != 0 {
		result["SecurityAudits"] = r.SecurityAudits
	}
	if r.IsVisible&(1<<4) != 0 {
		result["DataProtection"] = r.DataProtection
	}
	if r.IsVisible&(1<<5) != 0 {
		result["SecurityIncidents"] = r.SecurityIncidents
	}
	if r.IsVisible&(1<<6) != 0 {
		result["KeyDependencies"] = r.KeyDependencies
	}
	if r.IsVisible&(1<<7) != 0 {
		result["ContingencyPlans"] = r.ContingencyPlans
	}
	return result
}

func (r *RiskManagement) EditableFilter() error {
	var errFields []string

	if r.IsEditable&(1<<0) == 0 {
		errFields = append(errFields, "RegulatoryChanges")
	}
	if r.IsEditable&(1<<1) == 0 {
		errFields = append(errFields, "ComplianceStatus")
	}
	if r.IsEditable&(1<<2) == 0 {
		errFields = append(errFields, "RegulatoryConcerns")
	}
	if r.IsEditable&(1<<3) == 0 {
		errFields = append(errFields, "SecurityAudits")
	}
	if r.IsEditable&(1<<4) == 0 {
		errFields = append(errFields, "DataProtection")
	}
	if r.IsEditable&(1<<5) == 0 {
		errFields = append(errFields, "SecurityIncidents")
	}
	if r.IsEditable&(1<<6) == 0 {
		errFields = append(errFields, "KeyDependencies")
	}
	if r.IsEditable&(1<<7) == 0 {
		errFields = append(errFields, "ContingencyPlans")
	}

	if len(errFields) > 0 {
		return fmt.Errorf("fields not editable: %v", errFields)
	}
	return nil
}
